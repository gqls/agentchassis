-- ============================================================================
-- MIGRATION: Entity State Log & Builder Agents
-- ============================================================================
-- This migration:
-- 1. Creates the entity_state_log table for persistent cross-orchestration data
-- 2. Adds briefing_questionnaire column to agent_definitions
-- 3. Creates builder agent definitions (replacing agent_group_definitions)
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. ENTITY STATE LOG TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS entity_state_log (
                                                id BIGSERIAL PRIMARY KEY,

    -- Entity identification
                                                entity_id VARCHAR(255) NOT NULL,
    entity_type VARCHAR(100),           -- 'domain', 'project', 'customer', etc.
    namespace VARCHAR(100),             -- NULL=shared, or agent_type, or custom

-- Data
    path VARCHAR(255) NOT NULL,         -- 'brand.tone', 'research.products', etc.
    data JSONB NOT NULL,

    -- Context
    created_at TIMESTAMP DEFAULT now(),
    created_by_agent_type VARCHAR(100),
    orchestration_id UUID,
    correlation_id VARCHAR(100),

    -- For future intelligent supersession
    superseded_by BIGINT REFERENCES entity_state_log(id),
    supersession_reason TEXT
    );

-- Index for efficient lookups: get latest by entity + namespace + path
CREATE INDEX IF NOT EXISTS idx_entity_state_lookup
    ON entity_state_log(entity_id, namespace, path, created_at DESC);

-- Index for finding active (non-superseded) entries
CREATE INDEX IF NOT EXISTS idx_entity_state_active
    ON entity_state_log(entity_id, namespace, path)
    WHERE superseded_by IS NULL;

-- Index by entity type for bulk operations
CREATE INDEX IF NOT EXISTS idx_entity_state_type
    ON entity_state_log(entity_type, entity_id);

-- Index by orchestration for debugging/tracing
CREATE INDEX IF NOT EXISTS idx_entity_state_orchestration
    ON entity_state_log(orchestration_id);

-- Comments
COMMENT ON TABLE entity_state_log IS 'Append-only log of entity state changes, supporting persistent data across orchestrations';
COMMENT ON COLUMN entity_state_log.entity_id IS 'Identifier for the entity (e.g., domain name, project ID)';
COMMENT ON COLUMN entity_state_log.entity_type IS 'Type of entity: domain, project, customer, etc.';
COMMENT ON COLUMN entity_state_log.namespace IS 'NULL for shared data, agent_type for agent-specific data, or custom namespace';
COMMENT ON COLUMN entity_state_log.path IS 'Dot-notation path within namespace, e.g., brand.tone, research.products';
COMMENT ON COLUMN entity_state_log.superseded_by IS 'Points to newer entry that supersedes this one (for compaction)';

-- ============================================================================
-- 2. ADD NEW COLUMNS TO AGENT_DEFINITIONS
-- ============================================================================

-- Add briefing_questionnaire for intake/briefing workflows
ALTER TABLE agent_definitions
    ADD COLUMN IF NOT EXISTS briefing_questionnaire JSONB DEFAULT '{}'::jsonb;

-- Add usage_count for discovery ranking
ALTER TABLE agent_definitions
    ADD COLUMN IF NOT EXISTS usage_count INTEGER DEFAULT 0;

-- Add is_snapshot to mark frozen versions
ALTER TABLE agent_definitions
    ADD COLUMN IF NOT EXISTS is_snapshot BOOLEAN DEFAULT false;

-- Add index for type+version lookups
CREATE INDEX IF NOT EXISTS idx_agent_definitions_type_version
    ON agent_definitions(type, version DESC);

-- Add index for active agents ordered by usage
CREATE INDEX IF NOT EXISTS idx_agent_definitions_usage
    ON agent_definitions(usage_count DESC)
    WHERE is_active = true;

COMMENT ON COLUMN agent_definitions.briefing_questionnaire IS
'Optional questionnaire for briefing agents to execute when working with this agent type';

COMMENT ON COLUMN agent_definitions.usage_count IS
'Number of times this agent definition has been used, for discovery ranking';

COMMENT ON COLUMN agent_definitions.is_snapshot IS
'If true, this definition is frozen and should not be modified. Variants should reference snapshot versions.';

-- ============================================================================
-- 3. INTAKE ORCHESTRATOR AGENT
-- Entry point that classifies, briefs, and spawns the appropriate builder
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config,
    briefing_questionnaire
)
VALUES (
           gen_random_uuid(),
           'intake-orchestrator',
           'Intake Orchestrator',
           'Entry point for site creation: classifies project type, runs briefing, spawns appropriate builder agent',
           'orchestrator',
           '{
             "workflow": {
               "start_step": "spawn_classifier",
               "steps": {
                 "spawn_classifier": {
                   "action": "spawn_agent",
                   "config": {"role": "classifier", "agent_type": "site-classifier"},
                   "next_step": "spawn_briefer",
                   "description": "Spawn site classifier agent"
                 },
                 "spawn_briefer": {
                   "action": "spawn_agent",
                   "config": {"role": "briefer", "agent_type": "briefing-agent"},
                   "next_step": "call_classifier",
                   "description": "Spawn briefing agent"
                 },
                 "call_classifier": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "site-classifier",
                     "target_role": "classifier",
                     "timeout_seconds": 30
                   },
                   "output_field": "classification",
                   "next_step": "hitl_confirm_type",
                   "description": "Classify the site type from domain and objective"
                 },
                 "hitl_confirm_type": {
                   "action": "request_human_input",
                   "config": {
                     "request_type": "confirmation",
                     "title": "Confirm Site Type",
                     "message": "Please confirm or adjust the site classification",
                     "fields": [
                       {
                         "name": "site_type",
                         "type": "select",
                         "label": "Site Type",
                         "options": ["landing", "content", "portfolio", "brochure"],
                         "default_from": "classification.classify_site.result.site_type"
                       },
                       {
                         "name": "recommended_group",
                         "type": "select",
                         "label": "Builder",
                         "options": ["landing-page-builder", "content-site-builder", "portfolio-builder", "brochure-builder"],
                         "default_from": "classification.classify_site.result.recommended_group"
                       }
                     ],
                     "timeout_seconds": 86400,
                     "skip_if": "input_data.hitl_mode == auto"
                   },
                   "output_field": "confirmed_type",
                   "next_step": "fetch_questionnaire",
                   "description": "Human confirms or adjusts the site type classification"
                 },
                 "fetch_questionnaire": {
                   "action": "fetch_agent_questionnaire",
                   "config": {
                     "agent_type_field": "confirmed_type.recommended_group"
                   },
                   "output_field": "questionnaire",
                   "next_step": "call_briefer",
                   "description": "Fetch the briefing questionnaire for the target builder"
                 },
                 "call_briefer": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "briefing-agent",
                     "target_role": "briefer",
                     "input_fields": ["input_data", "classification", "confirmed_type", "questionnaire"],
                     "timeout_seconds": 120
                   },
                   "output_field": "brief_data",
                   "next_step": "hitl_review_brief",
                   "description": "Run the briefing questionnaire"
                 },
                 "hitl_review_brief": {
                   "action": "request_human_input",
                   "config": {
                     "request_type": "review",
                     "title": "Review Brief",
                     "message": "Please review and adjust the briefing answers if needed",
                     "data_field": "brief_data",
                     "editable": true,
                     "timeout_seconds": 86400,
                     "skip_if": "input_data.hitl_mode == auto"
                   },
                   "output_field": "reviewed_brief",
                   "next_step": "spawn_builder",
                   "description": "Human reviews and can edit the brief before proceeding"
                 },
                 "spawn_builder": {
                   "action": "spawn_agent",
                   "config": {
                     "agent_type_field": "confirmed_type.recommended_group",
                     "role": "builder",
                     "input_fields": ["input_data", "classification", "brief_data", "reviewed_brief"]
                   },
                   "output_field": "spawned_builder",
                   "next_step": "complete",
                   "description": "Spawn the appropriate builder agent with all collected data"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Intake complete - builder agent has been spawned"
                 }
               }
             },
             "processing_mode": "orchestration",
             "timeout_seconds": 600
           }'::jsonb,
           true,
           '["orchestration", "intake", "classification"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.478',
           '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
           '{}'::jsonb  -- Intake has no questionnaire - it fetches from target builder
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       updated_at = now();

-- ============================================================================
-- 4. LANDING PAGE BUILDER AGENT
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config,
    briefing_questionnaire
)
VALUES (
           gen_random_uuid(),
           'landing-page-builder',
           'Landing Page Builder',
           'Orchestrates the complete landing page build workflow - spawns specialist agents and coordinates them to build conversion-focused landing pages',
           'orchestrator',
           '{
             "workflow": {
               "start_step": "spawn_strategist",
               "steps": {
                 "spawn_strategist": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "site-strategist", "role": "strategist"},
                   "next_step": "spawn_architect",
                   "description": "Spawn strategist"
                 },
                 "spawn_architect": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "landing-page-architect", "role": "architect"},
                   "next_step": "spawn_writer",
                   "description": "Spawn landing page architect"
                 },
                 "spawn_writer": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "content-writer", "role": "writer"},
                   "next_step": "spawn_assembler",
                   "description": "Spawn content writer"
                 },
                 "spawn_assembler": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "html-assembler", "role": "assembler"},
                   "next_step": "spawn_wrapper",
                   "description": "Spawn HTML assembler"
                 },
                 "spawn_wrapper": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "multipage-wrapper", "role": "wrapper"},
                   "next_step": "spawn_deployer",
                   "description": "Spawn multipage wrapper"
                 },
                 "spawn_deployer": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "site-deployer", "role": "deployer"},
                   "next_step": "call_strategist",
                   "description": "Spawn deployer"
                 },
                 "call_strategist": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "site-strategist",
                     "target_role": "strategist",
                     "input_fields": ["input_data", "brief_data"],
                     "timeout_seconds": 120
                   },
                   "output_field": "build_plan",
                   "next_step": "call_architect",
                   "description": "Generate build plan from brief"
                 },
                 "call_architect": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "landing-page-architect",
                     "target_role": "architect",
                     "input_fields": ["build_plan", "brief_data", "input_data"],
                     "timeout_seconds": 120
                   },
                   "output_field": "template_data",
                   "next_step": "call_writer",
                   "description": "Assemble page template from components"
                 },
                 "call_writer": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "content-writer",
                     "target_role": "writer",
                     "input_fields": ["template_data", "build_plan", "brief_data", "input_data"],
                     "timeout_seconds": 300
                   },
                   "output_field": "content_data",
                   "next_step": "call_assembler",
                   "description": "Generate content for template placeholders"
                 },
                 "call_assembler": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "html-assembler",
                     "target_role": "assembler",
                     "input_fields": ["content_data", "template_data", "brief_data", "input_data"],
                     "timeout_seconds": 120
                   },
                   "output_field": "final_html",
                   "next_step": "call_wrapper",
                   "description": "Assemble final HTML with CSS/JS"
                 },
                 "call_wrapper": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "multipage-wrapper",
                     "target_role": "wrapper",
                     "input_fields": ["final_html", "input_data"],
                     "timeout_seconds": 60
                   },
                   "output_field": "site_files",
                   "next_step": "call_deployer",
                   "description": "Create about and contact pages, package as files map"
                 },
                 "call_deployer": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "site-deployer",
                     "target_role": "deployer",
                     "input_fields": ["site_files", "input_data"],
                     "timeout_seconds": 180
                   },
                   "output_field": "deployment_result",
                   "next_step": "complete",
                   "description": "Deploy to git repository"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Landing page build complete"
                 }
               }
             },
             "processing_mode": "orchestration",
             "timeout_seconds": 900
           }'::jsonb,
           true,
           '["orchestration", "site-building", "landing-page"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.478',
           '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
           '{
             "sections": [
               {
                 "name": "brand",
                 "title": "Brand & Identity",
                 "questions": [
                   {"field": "brand_name", "type": "text", "label": "Brand/Company Name", "required": true},
                   {"field": "tagline", "type": "text", "label": "Tagline or Slogan", "required": false},
                   {"field": "tone", "type": "select", "label": "Brand Tone", "options": ["professional", "friendly", "bold", "playful", "technical"], "default": "professional"}
                 ]
               },
               {
                 "name": "value_proposition",
                 "title": "Value Proposition",
                 "questions": [
                   {"field": "primary_benefit", "type": "textarea", "label": "What is the main benefit for visitors?", "required": true},
                   {"field": "unique_selling_points", "type": "textarea", "label": "What makes you different? (List 3-5 points)", "required": true},
                   {"field": "target_audience", "type": "text", "label": "Who is your ideal customer?", "required": true}
                 ]
               },
               {
                 "name": "conversion",
                 "title": "Conversion Goals",
                 "questions": [
                   {"field": "primary_cta", "type": "text", "label": "Primary Call-to-Action (e.g., Sign Up, Buy Now)", "required": true},
                   {"field": "primary_cta_url", "type": "text", "label": "Primary CTA Link/Action", "required": false},
                   {"field": "secondary_cta", "type": "text", "label": "Secondary CTA (e.g., Learn More)", "required": false}
                 ]
               },
               {
                 "name": "social_proof",
                 "title": "Trust & Social Proof",
                 "questions": [
                   {"field": "has_testimonials", "type": "boolean", "label": "Do you have customer testimonials?", "default": false},
                   {"field": "client_count", "type": "text", "label": "Number of customers/users (if applicable)", "required": false},
                   {"field": "notable_clients", "type": "text", "label": "Notable clients or partners", "required": false}
                 ]
               }
             ]
           }'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       briefing_questionnaire = EXCLUDED.briefing_questionnaire,
                                       description = EXCLUDED.description,
                                       updated_at = now();

-- ============================================================================
-- 4. CONTENT SITE BUILDER AGENT
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config,
    briefing_questionnaire
)
VALUES (
           gen_random_uuid(),
           'content-site-builder',
           'Content Site Builder',
           'Orchestrates the complete content/publishing site build workflow',
           'orchestrator',
           '{
             "workflow": {
               "start_step": "spawn_strategist",
               "steps": {
                 "spawn_strategist": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "site-strategist", "role": "strategist"},
                   "next_step": "spawn_architect"
                 },
                 "spawn_architect": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "content-site-architect", "role": "architect"},
                   "next_step": "spawn_writer"
                 },
                 "spawn_writer": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "content-writer", "role": "writer"},
                   "next_step": "spawn_assembler"
                 },
                 "spawn_assembler": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "html-assembler", "role": "assembler"},
                   "next_step": "spawn_wrapper"
                 },
                 "spawn_wrapper": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "multipage-wrapper", "role": "wrapper"},
                   "next_step": "spawn_deployer"
                 },
                 "spawn_deployer": {
                   "action": "spawn_agent",
                   "config": {"agent_type": "site-deployer", "role": "deployer"},
                   "next_step": "call_strategist"
                 },
                 "call_strategist": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "site-strategist",
                     "target_role": "strategist",
                     "input_fields": ["input_data", "brief_data"],
                     "timeout_seconds": 120
                   },
                   "output_field": "build_plan",
                   "next_step": "call_architect"
                 },
                 "call_architect": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "content-site-architect",
                     "target_role": "architect",
                     "input_fields": ["build_plan", "brief_data", "input_data"],
                     "timeout_seconds": 120
                   },
                   "output_field": "template_data",
                   "next_step": "call_writer"
                 },
                 "call_writer": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "content-writer",
                     "target_role": "writer",
                     "input_fields": ["template_data", "build_plan", "brief_data", "input_data"],
                     "timeout_seconds": 300
                   },
                   "output_field": "content_data",
                   "next_step": "call_assembler"
                 },
                 "call_assembler": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "html-assembler",
                     "target_role": "assembler",
                     "input_fields": ["content_data", "template_data", "brief_data", "input_data"],
                     "timeout_seconds": 120
                   },
                   "output_field": "final_html",
                   "next_step": "call_wrapper"
                 },
                 "call_wrapper": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "multipage-wrapper",
                     "target_role": "wrapper",
                     "input_fields": ["final_html", "input_data"],
                     "timeout_seconds": 60
                   },
                   "output_field": "site_files",
                   "next_step": "call_deployer",
                   "description": "Create about and contact pages, package as files map"
                 },
                 "call_deployer": {
                   "action": "call_agent",
                   "config": {
                     "agent_type": "site-deployer",
                     "target_role": "deployer",
                     "input_fields": ["site_files", "input_data"],
                     "timeout_seconds": 180
                   },
                   "output_field": "deployment_result",
                   "next_step": "complete"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Content site build complete"
                 }
               }
             },
             "processing_mode": "orchestration",
             "timeout_seconds": 900
           }'::jsonb,
           true,
           '["orchestration", "site-building", "content-site"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.478',
           '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb,
           '{
             "sections": [
               {
                 "name": "publication",
                 "title": "Publication Identity",
                 "questions": [
                   {"field": "publication_name", "type": "text", "label": "Publication/Site Name", "required": true},
                   {"field": "tagline", "type": "text", "label": "Tagline", "required": false},
                   {"field": "editorial_tone", "type": "select", "label": "Editorial Tone", "options": ["news_formal", "magazine_polished", "blog_casual", "technical"], "default": "magazine_polished"}
                 ]
               },
               {
                 "name": "content_structure",
                 "title": "Content Structure",
                 "questions": [
                   {"field": "categories", "type": "textarea", "label": "Content Categories (one per line)", "required": true},
                   {"field": "publishing_frequency", "type": "select", "label": "Publishing Frequency", "options": ["daily", "weekly", "occasional"], "default": "weekly"}
                 ]
               },
               {
                 "name": "monetization",
                 "title": "Monetization",
                 "questions": [
                   {"field": "monetization_model", "type": "select", "label": "Revenue Model", "options": ["advertising", "subscription", "affiliate", "none"], "default": "advertising"},
                   {"field": "newsletter_signup", "type": "boolean", "label": "Include Newsletter Signup?", "default": true}
                 ]
               }
             ]
           }'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       briefing_questionnaire = EXCLUDED.briefing_questionnaire,
                                       updated_at = now();

-- ============================================================================
-- 6. MULTIPAGE WRAPPER AGENT (may already exist)
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config
)
VALUES (
           gen_random_uuid(),
           'multipage-wrapper',
           'Multi-Page Site Wrapper',
           'Wraps single-page site into multi-page structure (index, about, contact)',
           'data-driven',
           '{
             "processing_mode": "task",
             "timeout_seconds": 30,
             "workflow": {
               "start_step": "wrap_multipage",
               "steps": {
                 "wrap_multipage": {
                   "action": "wrap_multipage",
                   "config": {
                     "index_html_field": "input_data.final_html.assemble_html.final_html"
                   },
                   "next_step": "complete",
                   "description": "Create about and contact pages"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Return files map"
                 }
               }
             }
           }'::jsonb,
           true,
           '["data-transformation", "html", "multipage"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.478',
           '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       updated_at = now();

-- ============================================================================
-- 7. RELATIONSHIPS TABLE
-- First-class relationships between entities (roles, agents, external parties)
-- ============================================================================

CREATE TABLE IF NOT EXISTS relationships (
                                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Endpoints
    source_entity_id VARCHAR(255) NOT NULL,
    source_entity_type VARCHAR(100) NOT NULL,  -- 'role', 'agent', 'external', 'position'
    target_entity_id VARCHAR(255) NOT NULL,
    target_entity_type VARCHAR(100) NOT NULL,

    -- Relationship properties
    relationship_type VARCHAR(100) NOT NULL,   -- 'reports_to', 'collaborates_with', 'supplies_to', etc.
    direction VARCHAR(20) DEFAULT 'one_way',   -- 'one_way', 'bidirectional'

-- Relationship-specific configuration
    properties JSONB DEFAULT '{}'::jsonb,      -- communication preferences, protocols, etc.

-- State
    status VARCHAR(50) DEFAULT 'active',       -- 'active', 'strained', 'dormant', 'ended'

-- Lifecycle
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    ended_at TIMESTAMP,

    -- Ensure unique relationships
    CONSTRAINT unique_relationship UNIQUE (source_entity_id, source_entity_type, target_entity_id, target_entity_type, relationship_type)
    );

-- Index for finding all relationships for an entity
CREATE INDEX IF NOT EXISTS idx_relationships_source
    ON relationships(source_entity_id, source_entity_type);

CREATE INDEX IF NOT EXISTS idx_relationships_target
    ON relationships(target_entity_id, target_entity_type);

-- Index for finding relationships by type
CREATE INDEX IF NOT EXISTS idx_relationships_type
    ON relationships(relationship_type);

-- Index for active relationships only
CREATE INDEX IF NOT EXISTS idx_relationships_active
    ON relationships(source_entity_id, target_entity_id)
    WHERE status = 'active';

COMMENT ON TABLE relationships IS 'First-class relationships between entities, with their own identity and state';
COMMENT ON COLUMN relationships.source_entity_id IS 'ID of the source entity in the relationship';
COMMENT ON COLUMN relationships.source_entity_type IS 'Type of source: role, agent, external, position';
COMMENT ON COLUMN relationships.relationship_type IS 'Nature of relationship: reports_to, collaborates_with, etc.';
COMMENT ON COLUMN relationships.properties IS 'Relationship-specific config: communication preferences, protocols';
COMMENT ON COLUMN relationships.status IS 'Current state: active, strained, dormant, ended';

-- ============================================================================
-- 8. VERIFY MIGRATION
-- ============================================================================

DO $$
BEGIN
    -- Check entity_state_log table exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'entity_state_log') THEN
        RAISE EXCEPTION 'entity_state_log table was not created';
END IF;

    -- Check relationships table exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'relationships') THEN
        RAISE EXCEPTION 'relationships table was not created';
END IF;

    -- Check new columns on agent_definitions
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'agent_definitions' AND column_name = 'usage_count') THEN
        RAISE EXCEPTION 'usage_count column was not added to agent_definitions';
END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                   WHERE table_name = 'agent_definitions' AND column_name = 'is_snapshot') THEN
        RAISE EXCEPTION 'is_snapshot column was not added to agent_definitions';
END IF;

    -- Check intake-orchestrator agent exists
    IF NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'intake-orchestrator') THEN
        RAISE EXCEPTION 'intake-orchestrator agent was not created';
END IF;

    -- Check builder agents exist
    IF NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'landing-page-builder') THEN
        RAISE EXCEPTION 'landing-page-builder agent was not created';
END IF;

    IF NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'content-site-builder') THEN
        RAISE EXCEPTION 'content-site-builder agent was not created';
END IF;

    IF NOT EXISTS (SELECT 1 FROM agent_definitions WHERE type = 'multipage-wrapper') THEN
        RAISE EXCEPTION 'multipage-wrapper agent was not created';
END IF;

    RAISE NOTICE 'Migration completed successfully';
END $$;

COMMIT;

===


adding improvement proposals table for when using discovery_actions
-- ============================================================================
-- 8. IMPROVEMENT PROPOSALS TABLE
-- Queue of proposed improvements awaiting HITL review
-- ============================================================================

CREATE TABLE IF NOT EXISTS improvement_proposals (
                                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- What is being improved
    target_type VARCHAR(50) NOT NULL,         -- 'agent_definition', 'variant', 'entity'
    target_id VARCHAR(255) NOT NULL,          -- agent type, variant ID, or entity ID

-- The proposal
    proposed_changes JSONB NOT NULL,          -- analysis, suggestions, specific changes

-- Source of the proposal
    source VARCHAR(50) NOT NULL,              -- 'metrics', 'agent_observation', 'human'
    source_agent_type VARCHAR(100),           -- if proposed by an agent
    source_orchestration_id UUID,             -- originating orchestration

-- Review status
    status VARCHAR(50) DEFAULT 'pending',     -- 'pending', 'approved', 'rejected', 'applied'
    reviewed_by VARCHAR(255),                 -- user or agent that reviewed
    reviewed_at TIMESTAMP,
    review_notes TEXT,

    -- Applied changes (if approved)
    applied_changes JSONB,                    -- what was actually applied
    applied_at TIMESTAMP,

    -- Lifecycle
    created_at TIMESTAMP DEFAULT now(),
    expires_at TIMESTAMP,                     -- optional expiry for time-sensitive proposals

-- For grouping related proposals
    correlation_id VARCHAR(100)
    );

-- Index for pending proposals
CREATE INDEX IF NOT EXISTS idx_improvement_proposals_pending
    ON improvement_proposals(target_type, target_id, status)
    WHERE status = 'pending';

-- Index for finding proposals by source
CREATE INDEX IF NOT EXISTS idx_improvement_proposals_source
    ON improvement_proposals(source, created_at DESC);

-- Index for reviewing proposals
CREATE INDEX IF NOT EXISTS idx_improvement_proposals_review
    ON improvement_proposals(status, created_at ASC)
    WHERE status = 'pending';

COMMENT ON TABLE improvement_proposals IS 'Queue of proposed improvements to agents, variants, or entities, awaiting human review';
COMMENT ON COLUMN improvement_proposals.target_type IS 'What type of thing is being improved';
COMMENT ON COLUMN improvement_proposals.source IS 'How the proposal was generated: metrics analysis, agent observation, or human suggestion';
COMMENT ON COLUMN improvement_proposals.status IS 'pending=awaiting review, approved=will apply, rejected=declined, applied=done';

----
additions - changes to accommodate deprecation of agent_group_definitions
and improvement of site classifier to allow for dynamic choice of site builder

-- ============================================================================
-- UPDATE INTAKE ORCHESTRATOR WITH DYNAMIC BUILDER DISCOVERY
-- ============================================================================
--
-- This update makes the intake-orchestrator self-describing:
-- 1. Queries available builders from agent_definitions
-- 2. Passes them to the classifier so it knows what's available
-- 3. Uses dynamic spawn to call the recommended builder
--
-- Prerequisites:
-- - query_agent_definitions action deployed
-- - spawn_agent dynamic field resolution patch deployed
-- ============================================================================

BEGIN;

-- ============================================================================
-- 1. UPDATE INTAKE-ORCHESTRATOR WORKFLOW
-- ============================================================================

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = '{
      "workflow": {
        "start_step": "fetch_available_builders",
        "steps": {
          "fetch_available_builders": {
            "action": "query_agent_definitions",
            "config": {
              "filter": {
                "type_pattern": "%-builder"
              },
              "fields": ["type", "display_name", "description"]
            },
            "output_field": "available_builders",
            "next_step": "spawn_classifier",
            "description": "Discover what builder agents are available"
          },

          "spawn_classifier": {
            "action": "spawn_agent",
            "config": {"role": "classifier", "agent_type": "site-classifier"},
            "next_step": "spawn_briefer",
            "description": "Spawn site classifier agent"
          },

          "spawn_briefer": {
            "action": "spawn_agent",
            "config": {"role": "briefer", "agent_type": "briefing-agent"},
            "next_step": "call_classifier",
            "description": "Spawn briefing agent"
          },

          "call_classifier": {
            "action": "call_agent",
            "config": {
              "agent_type": "site-classifier",
              "target_role": "classifier",
              "input_fields": ["input_data", "available_builders"],
              "timeout_seconds": 30
            },
            "output_field": "classification",
            "next_step": "hitl_confirm_type",
            "description": "Classify the site type from domain and objective"
          },

          "hitl_confirm_type": {
            "action": "request_human_input",
            "config": {
              "request_type": "confirmation",
              "title": "Confirm Site Type",
              "message": "Please confirm or adjust the site classification",
              "fields": [
                {
                  "name": "site_type",
                  "type": "select",
                  "label": "Site Type",
                  "options": ["landing", "content", "portfolio", "brochure"],
                  "default_from": "classification.classify_site.result.site_type"
                },
                {
                  "name": "recommended_builder",
                  "type": "dynamic_select",
                  "label": "Builder",
                  "options_from": "available_builders.agents",
                  "option_value_field": "type",
                  "option_label_field": "display_name",
                  "default_from": "classification.classify_site.result.recommended_builder"
                }
              ],
              "timeout_seconds": 86400,
              "skip_if": "input_data.hitl_mode == auto"
            },
            "output_field": "confirmed_type",
            "next_step": "fetch_questionnaire",
            "description": "Human confirms or adjusts the site type classification"
          },

          "fetch_questionnaire": {
            "action": "fetch_agent_questionnaire",
            "config": {
              "agent_type_field": "confirmed_type.recommended_builder"
            },
            "output_field": "questionnaire",
            "next_step": "call_briefer",
            "description": "Fetch the briefing questionnaire for the target builder"
          },

          "call_briefer": {
            "action": "call_agent",
            "config": {
              "agent_type": "briefing-agent",
              "target_role": "briefer",
              "input_fields": ["input_data", "classification", "confirmed_type", "questionnaire"],
              "timeout_seconds": 120
            },
            "output_field": "brief_data",
            "next_step": "hitl_review_brief",
            "description": "Run the briefing questionnaire"
          },

          "hitl_review_brief": {
            "action": "request_human_input",
            "config": {
              "request_type": "review",
              "title": "Review Brief",
              "message": "Please review and adjust the briefing answers if needed",
              "data_field": "brief_data",
              "editable": true,
              "timeout_seconds": 86400,
              "skip_if": "input_data.hitl_mode == auto"
            },
            "output_field": "reviewed_brief",
            "next_step": "spawn_builder",
            "description": "Human reviews the completed brief"
          },

          "spawn_builder": {
            "action": "spawn_agent",
            "config": {
              "agent_type_field": "confirmed_type.recommended_builder",
              "role": "builder",
              "input_fields": ["input_data", "classification", "brief_data", "reviewed_brief"]
            },
            "output_field": "spawned_builder",
            "next_step": "complete",
            "description": "Spawn the appropriate builder agent with all collected data"
          },

          "complete": {
            "action": "complete_workflow",
            "description": "Intake complete - builder has been spawned"
          }
        }
      },
      "processing_mode": "orchestration",
      "timeout_seconds": 600
    }'::jsonb
WHERE type = 'intake-orchestrator';


-- ============================================================================
-- 2. UPDATE SITE-CLASSIFIER TO USE AVAILABLE BUILDERS
-- ============================================================================

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = '{
      "workflow": {
        "start_step": "classify_site",
        "steps": {
          "classify_site": {
            "action": "execute_llm_prompt",
            "config": {
              "ai_service": {
                "provider": "anthropic",
                "model": "claude-haiku-4-5-20251001",
                "api_key_env_var": "ANTHROPIC_API_KEY",
                "max_tokens": 1500
              },
              "input_fields": ["input_data", "available_builders"],
              "output_field": "classification_result",
              "prompt_template": "Classify this website project and recommend the appropriate builder.\n\nInput:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n\nAvailable Builders:\n{{range .available_builders.agents}}- {{.type}}: {{.description}}\n{{end}}\n\nClassify the site into ONE of these types based on the objective:\n\n**landing** - Conversion-focused single-purpose sites:\n- Product/service sales pages, SaaS landing pages\n- Lead generation, signups, app downloads\n- Event registration, clear single CTA goal\n\n**content** - Publishing/content sites:\n- News, blogs, magazines, articles\n- Content aggregation, SEO/traffic focused\n- Category navigation, archives\n\n**portfolio** - Showcase/portfolio sites:\n- Creative portfolios, agencies, case studies\n- Visual/image heavy, project galleries\n\n**brochure** - Multi-page business sites:\n- Company websites with About, Services, Team, Contact\n- Informational focus\n\nAnalyze the domain name and stated objective to determine the best fit.\n\nReturn ONLY valid JSON:\n{\n  \"site_type\": \"landing|content|portfolio|brochure\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"Brief explanation of classification\",\n  \"recommended_builder\": \"<exact type from Available Builders list>\",\n  \"detected_industry\": \"Industry/niche if detectable\",\n  \"detected_signals\": [\"Signal 1\", \"Signal 2\"]\n}"
            },
            "next_step": "complete"
          },
          "complete": {
            "action": "complete_workflow",
            "description": "Return classification result"
          }
        }
      },
      "processing_mode": "task",
      "timeout_seconds": 30
    }'::jsonb
WHERE type = 'site-classifier';


-- ============================================================================
-- 3. VERIFY CHANGES
-- ============================================================================

DO $$
DECLARE
intake_start_step TEXT;
    classifier_input_fields JSONB;
BEGIN
    -- Check intake-orchestrator has fetch_available_builders as start
SELECT default_config->'workflow'->>'start_step' INTO intake_start_step
FROM agent_definitions WHERE type = 'intake-orchestrator';

IF intake_start_step != 'fetch_available_builders' THEN
        RAISE EXCEPTION 'intake-orchestrator start_step not updated. Got: %', intake_start_step;
END IF;

    -- Check classifier uses available_builders input
SELECT default_config->'workflow'->'steps'->'classify_site'->'config'->'input_fields' INTO classifier_input_fields
FROM agent_definitions WHERE type = 'site-classifier';

IF NOT classifier_input_fields @> '["available_builders"]'::jsonb THEN
        RAISE EXCEPTION 'site-classifier not using available_builders input';
END IF;

    RAISE NOTICE 'Intake flow updated successfully with dynamic builder discovery';
END $$;

COMMIT;


-- ============================================================================
-- NOTES
-- ============================================================================
--
-- Field name changes:
--   recommended_group -> recommended_builder
--
-- This is clearer because:
--   1. It describes what's being recommended (a builder agent)
--   2. It's not tied to the deprecated "group" concept
--
-- The flow now:
--   1. fetch_available_builders - queries agent_definitions for *-builder types
--   2. call_classifier - passes available_builders to LLM so it knows options
--   3. hitl_confirm_type - shows dynamic dropdown from available_builders
--   4. spawn_builder - uses agent_type_field to spawn the recommended one
--
-- To add a new builder:
--   1. INSERT into agent_definitions with type ending in '-builder'
--   2. It automatically appears in classifier options and HITL dropdown
--   3. No workflow changes needed

==
fix for site categorisation

-- ============================================================================
-- FIX: Update site-classifier to output recommended_builder instead of recommended_group
-- ============================================================================
--
-- The intake-orchestrator workflow expects:
--   classification.classify_site.result.recommended_builder
--
-- But the classifier was outputting:
--   classification.classify_site.result.recommended_group
--
-- This updates the classifier prompt to use the new field name.
-- ============================================================================

BEGIN;

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,classify_site,config,prompt_template}',
            to_jsonb(
                    'Classify this website project and recommend the appropriate builder.

        Input:
        - Domain: {{.input_data.domain}}
        - Objective: {{.input_data.objective}}

        Available Builders:
        {{range .available_builders.agents}}- {{.type}}: {{.description}}
        {{end}}

        Classify the site into ONE of these types based on the objective:

        **landing** - Conversion-focused single-purpose sites:
        - Product/service sales pages, SaaS landing pages
        - Lead generation, signups, app downloads
        - Event registration, clear single CTA goal

        **content** - Publishing/content sites:
        - News, blogs, magazines, articles
        - Content aggregation, SEO/traffic focused
        - Category navigation, archives

        **portfolio** - Showcase/portfolio sites:
        - Creative portfolios, agencies, case studies
        - Visual/image heavy, project galleries

        **brochure** - Multi-page business sites:
        - Company websites with About, Services, Team, Contact
        - Informational focus

        Analyze the domain name and stated objective to determine the best fit.

        Return ONLY valid JSON:
        {
          "site_type": "landing|content|portfolio|brochure",
          "confidence": 0.0-1.0,
          "reasoning": "Brief explanation of classification",
          "recommended_builder": "<exact type from Available Builders list>",
          "detected_industry": "Industry/niche if detectable",
          "detected_signals": ["Signal 1", "Signal 2"]
        }'
            )
                     )
WHERE type = 'site-classifier';

-- Also update the HITL confirmation fields to use recommended_builder
UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,hitl_confirm_type,config,fields}',
            '[
              {
                "name": "site_type",
                "type": "select",
                "label": "Site Type",
                "options": ["landing", "content", "portfolio", "brochure"],
                "default_from": "classification.classify_site.result.site_type"
              },
              {
                "name": "recommended_builder",
                "type": "dynamic_select",
                "label": "Builder",
                "options_from": "available_builders.agents",
                "option_value_field": "type",
                "option_label_field": "display_name",
                "default_from": "classification.classify_site.result.recommended_builder"
              }
            ]'::jsonb
                     )
WHERE type = 'intake-orchestrator';

-- Verify the changes
DO $$
DECLARE
classifier_prompt TEXT;
    intake_fields JSONB;
BEGIN
    -- Check classifier prompt contains recommended_builder
SELECT default_config->'workflow'->'steps'->'classify_site'->'config'->>'prompt_template'
INTO classifier_prompt
FROM agent_definitions WHERE type = 'site-classifier';

IF classifier_prompt NOT LIKE '%recommended_builder%' THEN
        RAISE EXCEPTION 'Classifier prompt not updated - still using old field name';
END IF;

    IF classifier_prompt LIKE '%recommended_group%' THEN
        RAISE EXCEPTION 'Classifier prompt still contains recommended_group';
END IF;

    -- Check intake fields use recommended_builder
SELECT default_config->'workflow'->'steps'->'hitl_confirm_type'->'config'->'fields'
INTO intake_fields
FROM agent_definitions WHERE type = 'intake-orchestrator';

IF intake_fields::text NOT LIKE '%recommended_builder%' THEN
        RAISE EXCEPTION 'Intake HITL fields not updated';
END IF;

    RAISE NOTICE 'Successfully updated classifier and intake to use recommended_builder';
END $$;

COMMIT;

--
update cleaner

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = jsonb_set(
            default_config,
            '{workflow,steps,classify_site,config,prompt_template}',
            '"Classify this website project and recommend the appropriate builder.\n\nInput:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n\nAvailable Builders:\n{{range .available_builders.agents}}- {{.type}}: {{.description}}\n{{end}}\n\nClassify the site into ONE of these types based on the objective:\n\n**landing** - Conversion-focused single-purpose sites:\n- Product/service sales pages, SaaS landing pages\n- Lead generation, signups, app downloads\n- Event registration, clear single CTA goal\n\n**content** - Publishing/content sites:\n- News, blogs, magazines, articles\n- Content aggregation, SEO/traffic focused\n- Category navigation, archives\n\n**portfolio** - Showcase/portfolio sites:\n- Creative portfolios, agencies, case studies\n- Visual/image heavy, project galleries\n\n**brochure** - Multi-page business sites:\n- Company websites with About, Services, Team, Contact\n- Informational focus\n\nAnalyze the domain name and stated objective to determine the best fit.\n\nReturn ONLY valid JSON:\n{\n  \"site_type\": \"landing|content|portfolio|brochure\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"Brief explanation of classification\",\n  \"recommended_builder\": \"<exact type from Available Builders list>\",\n  \"detected_industry\": \"Industry/niche if detectable\",\n  \"detected_signals\": [\"Signal 1\", \"Signal 2\"]\n}"'::jsonb
                     )
WHERE type = 'site-classifier';

