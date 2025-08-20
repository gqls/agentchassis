-- FILE: platform/database/migrations/005_website_builder_agents.sql
-- Complete website builder agents configuration with all supporting tables

-- ============================================================================
-- PART 1: AGENT DEFINITIONS (All 8 agents already defined)
-- ============================================================================

-- 1. Domain Analyst (with full config preserved)
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, command,
    resources, default_config, capabilities, topics, health_config
) VALUES (
             'domain-analyst',
             'Domain Analyst',
             'Analyzes domains and determines appropriate website type',
             'data-driven',
             'docker.io/aqls/agent-chassis',
             'v1.0.48',
             ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
             '{
                 "requests": {"cpu": "100m", "memory": "256Mi"},
                 "limits": {"cpu": "500m", "memory": "512Mi"}
             }'::jsonb,
             '{
                 "model": "claude-3-5-sonnet-20241022",
                 "temperature": 0.3,
                 "processing_mode": "task",
                 "ai_service": {
                     "provider": "anthropic",
                     "model": "claude-3-5-sonnet-20241022"
                 },
                 "prompt_template": "You are a business domain expert. Analyze the following business name: ''{{.input.business_name}}'' and domain: ''{{.input.domain}}''. Based on this information, provide a structured JSON output with the business ''type'' (e.g., ''e-commerce'', ''portfolio'', ''blog'') and a list of 5 relevant SEO ''keywords''.",
                 "workflow": {
                     "start_step": "analyze",
                     "steps": {
                         "analyze": {
                             "action": "execute_llm_prompt",
                             "description": "Analyze the domain",
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Complete the analysis"
                         }
                     }
                 }
             }'::jsonb,
             '["analysis", "categorization", "domain-research"]'::jsonb,
             '{
                 "process": "system.agent.domain-analyst.process",
                 "response": "system.responses.domain-analyst",
                 "error": "system.errors.domain-analyst",
                 "dlq": "dlq.domain-analyst"
             }'::jsonb,
             '{
                 "liveness_path": "/health",
                 "readiness_path": "/ready",
                 "port": 8080,
                 "initial_delay_seconds": 30
             }'::jsonb
         ) ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    image_repository = EXCLUDED.image_repository,
    image_tag = EXCLUDED.image_tag,
    command = EXCLUDED.command,
    resources = EXCLUDED.resources,
    default_config = EXCLUDED.default_config,
    capabilities = EXCLUDED.capabilities,
    topics = EXCLUDED.topics,
    updated_at = NOW();

-- 2. Site Architect (with full config preserved)
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, command,
    resources, default_config, capabilities, topics
) VALUES (
             'site-architect',
             'Site Architect',
             'Plans website structure and navigation',
             'data-driven',
             'docker.io/aqls/agent-chassis',
             'v1.0.48',
             ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
             '{
                 "requests": {"cpu": "100m", "memory": "256Mi"},
                 "limits": {"cpu": "500m", "memory": "512Mi"}
             }'::jsonb,
             '{
                 "model": "claude-3-5-sonnet-20241022",
                 "temperature": 0.5,
                 "processing_mode": "task",
                 "ai_service": {
                     "provider": "anthropic",
                     "model": "claude-3-5-sonnet-20241022"
                 },
                 "prompt_template": "You are a website architect. For a ''{{.input.business_type}}'' business named ''{{.input.business_name}}'', design a complete site structure. Provide a structured JSON output with a ''pages'' array. Each object in the array should have a ''name'' (e.g., ''Home'', ''About Us'') and a brief ''description'' of its purpose.",
                 "workflow": {
                     "start_step": "design_structure",
                     "steps": {
                         "design_structure": {
                             "action": "execute_llm_prompt",
                             "description": "Design site structure",
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow"
                         }
                     }
                 }
             }'::jsonb,
             '["planning", "structure", "navigation"]'::jsonb,
             '{
                 "process": "system.agent.site-architect.process",
                 "response": "system.responses.site-architect",
                 "error": "system.errors.site-architect",
                 "dlq": "dlq.site-architect"
             }'::jsonb
         ) ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    image_repository = EXCLUDED.image_repository,
    image_tag = EXCLUDED.image_tag,
    command = EXCLUDED.command,
    resources = EXCLUDED.resources,
    default_config = EXCLUDED.default_config,
    capabilities = EXCLUDED.capabilities,
    topics = EXCLUDED.topics,
    updated_at = NOW();

-- 3. Content Researcher (with FULL complex workflow preserved)
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, command,
    resources, default_config, capabilities, topics
) VALUES (
             'content-researcher',
             'Content Researcher',
             'Researches and gathers comprehensive information for website content',
             'data-driven',
             'docker.io/aqls/agent-chassis',
             'v1.0.48',
             ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
             '{
                 "requests": {"cpu": "150m", "memory": "384Mi"},
                 "limits": {"cpu": "600m", "memory": "768Mi"}
             }'::jsonb,
             '{
                 "model": "claude-3-5-sonnet-20241022",
                 "temperature": 0.4,
                 "processing_mode": "task",
                 "ai_service": {
                     "provider": "anthropic",
                     "model": "claude-3-5-sonnet-20241022"
                 },
                 "workflow": {
                     "start_step": "identify_topics",
                     "steps": {
                         "identify_topics": {
                             "action": "validate_input",
                             "description": "Identify research topics",
                             "next_step": "deep_research"
                         },
                         "deep_research": {
                             "action": "call_agent",
                             "description": "Perform deep research using Perplexity",
                             "agent_type": "perplexity-research",
                             "topic": "system.adapter.perplexity.research",
                             "config": {
                                 "mode": "comprehensive",
                                 "include_sources": true,
                                 "max_depth": 3
                             },
                             "next_step": "crawl_competitors"
                         },
                         "crawl_competitors": {
                             "action": "call_agent",
                             "description": "Analyze competitor websites",
                             "agent_type": "firecrawl-scraper",
                             "topic": "system.adapter.firecrawl.scrape",
                             "config": {
                                 "scrape_type": "competitor_analysis",
                                 "extract_content": true,
                                 "extract_meta": true
                             },
                             "next_step": "analyze_findings"
                         },
                         "analyze_findings": {
                             "action": "execute_llm_prompt",
                             "description": "Synthesize research findings",
                             "config": {"prompt_template": "synthesize_research"},
                             "store_memory": true,
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Complete research workflow"
                         }
                     }
                 }
             }'::jsonb,
             '["research", "analysis", "fact-checking", "content-gathering", "competitor-analysis"]'::jsonb,
             '{
                 "process": "system.agent.content-researcher.process",
                 "response": "system.responses.content-researcher",
                 "error": "system.errors.content-researcher",
                 "dlq": "dlq.content-researcher"
             }'::jsonb
         ) ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    image_repository = EXCLUDED.image_repository,
    image_tag = EXCLUDED.image_tag,
    command = EXCLUDED.command,
    resources = EXCLUDED.resources,
    default_config = EXCLUDED.default_config,
    capabilities = EXCLUDED.capabilities,
    topics = EXCLUDED.topics,
    updated_at = NOW();

-- 4. HTML Developer (with full config preserved)
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, command,
    resources, default_config, capabilities, topics
) VALUES (
             'html-developer',
             'HTML Developer',
             'Generates HTML/CSS/JS code for websites',
             'code-driven',
             'docker.io/aqls/agent-chassis',
             'v1.0.48',
             ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
             '{
                 "requests": {"cpu": "150m", "memory": "384Mi"},
                 "limits": {"cpu": "800m", "memory": "1Gi"}
             }'::jsonb,
             '{
                 "model": "claude-3-5-sonnet-20241022",
                 "temperature": 0.2,
                 "processing_mode": "task",
                 "ai_service": {
                     "provider": "anthropic",
                     "model": "claude-3-5-sonnet-20241022"
                 },
                 "prompt_template": "You are an expert front-end web developer. Generate a single, complete, and responsive HTML file (including inline CSS in a <style> tag) for the ''{{.input.page.name}}'' page of a website for ''{{.input.business_name}}''. The page''s purpose is: {{.input.page.description}}. Use the following content: {{.input.page.content}}.",
                 "workflow": {
                     "start_step": "receive_specs",
                     "steps": {
                         "receive_specs": {
                             "action": "validate_input",
                             "description": "Validate page specifications",
                             "next_step": "generate_template"
                         },
                         "generate_template": {
                             "action": "execute_llm_prompt",
                             "description": "Generate HTML template",
                             "config": {
                                 "language": "html",
                                 "framework": "vanilla",
                                 "responsive": true
                             },
                             "next_step": "create_pages"
                         },
                         "create_pages": {
                             "action": "fan_out",
                             "description": "Create all pages in parallel",
                             "config": {"per_page": true},
                             "next_step": "bundle_site"
                         },
                         "bundle_site": {
                             "action": "transform_data",
                             "description": "Bundle all files together",
                             "config": {"transformation": "bundle_files"},
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Complete HTML generation"
                         }
                     }
                 }
             }'::jsonb,
             '["html", "css", "javascript", "frontend"]'::jsonb,
             '{
                 "process": "system.agent.html-developer.process",
                 "response": "system.responses.html-developer",
                 "error": "system.errors.html-developer",
                 "dlq": "dlq.html-developer"
             }'::jsonb
         ) ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    image_repository = EXCLUDED.image_repository,
    image_tag = EXCLUDED.image_tag,
    command = EXCLUDED.command,
    resources = EXCLUDED.resources,
    default_config = EXCLUDED.default_config,
    capabilities = EXCLUDED.capabilities,
    topics = EXCLUDED.topics,
    updated_at = NOW();

-- 5. Visual Designer (with full Firecrawl workflow preserved)
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, command,
    resources, default_config, capabilities, topics
) VALUES (
             'visual-designer',
             'Visual Designer',
             'Handles images, logos, and visual assets',
             'adapter',
             'docker.io/aqls/agent-chassis',
             'v1.0.48',
             ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
             '{
                 "requests": {"cpu": "100m", "memory": "256Mi"},
                 "limits": {"cpu": "500m", "memory": "768Mi"}
             }'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "workflow": {
                     "start_step": "analyze_brand",
                     "steps": {
                         "analyze_brand": {
                             "action": "validate_input",
                             "description": "Analyze brand requirements",
                             "next_step": "search_images"
                         },
                         "search_images": {
                             "action": "call_agent",
                             "description": "Search for relevant images",
                             "agent_type": "firecrawl-scraper",
                             "topic": "system.adapter.firecrawl.scrape",
                             "config": {
                                 "scrape_type": "image_search",
                                 "sources": ["unsplash", "pexels"],
                                 "extract_images": true
                             },
                             "next_step": "create_logo"
                         },
                         "create_logo": {
                             "action": "call_agent",
                             "description": "Generate logo using AI",
                             "agent_type": "image-generator",
                             "topic": "system.adapter.image.generate",
                             "next_step": "optimize_images"
                         },
                         "optimize_images": {
                             "action": "transform_data",
                             "description": "Optimize all images",
                             "config": {"transformation": "optimize_images"},
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Complete visual design workflow"
                         }
                     }
                 }
             }'::jsonb,
             '["design", "graphics", "branding", "image-processing"]'::jsonb,
             '{
                 "process": "system.agent.visual-designer.process",
                 "response": "system.responses.visual-designer",
                 "error": "system.errors.visual-designer",
                 "dlq": "dlq.visual-designer"
             }'::jsonb
         ) ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    image_repository = EXCLUDED.image_repository,
    image_tag = EXCLUDED.image_tag,
    command = EXCLUDED.command,
    resources = EXCLUDED.resources,
    default_config = EXCLUDED.default_config,
    capabilities = EXCLUDED.capabilities,
    topics = EXCLUDED.topics,
    updated_at = NOW();

-- 6. Site Publisher (with full workflow preserved)
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, command,
    resources, default_config, capabilities, topics
) VALUES (
             'site-publisher',
             'Site Publisher',
             'Publishes websites to storage buckets',
             'adapter',
             'docker.io/aqls/agent-chassis',
             'v1.0.48',
             ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
             '{
                 "requests": {"cpu": "100m", "memory": "256Mi"},
                 "limits": {"cpu": "400m", "memory": "512Mi"}
             }'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "workflow": {
                     "start_step": "collect_assets",
                     "steps": {
                         "collect_assets": {
                             "action": "validate_input",
                             "description": "Collect all website assets",
                             "next_step": "organize_files"
                         },
                         "organize_files": {
                             "action": "transform_data",
                             "description": "Organize file structure",
                             "config": {"transformation": "organize_structure"},
                             "next_step": "upload_to_bucket"
                         },
                         "upload_to_bucket": {
                             "action": "s3_upload",
                             "description": "Upload to S3 bucket",
                             "config": {
                                 "bucket": "${SITE_BUCKET}",
                                 "public": true
                             },
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Complete publishing workflow"
                         }
                     }
                 }
             }'::jsonb,
             '["deployment", "hosting", "publishing", "s3"]'::jsonb,
             '{
                 "process": "system.agent.site-publisher.process",
                 "response": "system.responses.site-publisher",
                 "error": "system.errors.site-publisher",
                 "dlq": "dlq.site-publisher"
             }'::jsonb
         ) ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    image_repository = EXCLUDED.image_repository,
    image_tag = EXCLUDED.image_tag,
    command = EXCLUDED.command,
    resources = EXCLUDED.resources,
    default_config = EXCLUDED.default_config,
    capabilities = EXCLUDED.capabilities,
    topics = EXCLUDED.topics,
    updated_at = NOW();

-- 7. Website Builder Orchestrator (with full complex workflow preserved)
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, command,
    resources, default_config, capabilities, topics
) VALUES (
             'website-builder',
             'Website Builder',
             'Orchestrates complete website creation',
             'orchestrator',
             'docker.io/aqls/agent-chassis',
             'v1.0.48',
             ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
             '{
                 "requests": {"cpu": "200m", "memory": "512Mi"},
                 "limits": {"cpu": "1000m", "memory": "1Gi"}
             }'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "workflow": {
                     "start_step": "validate_request",
                     "steps": {
                         "validate_request": {
                             "action": "validate_input",
                             "description": "Validate website creation request",
                             "next_step": "spawn_agents"
                         },
                         "spawn_agents": {
                             "action": "spawn_group",
                             "description": "Spawn the website builder team",
                             "config": {"group_type": "website-builder"},
                             "next_step": "analyze_domain"
                         },
                         "analyze_domain": {
                             "action": "call_agent",
                             "description": "Analyze the business domain",
                             "agent_type": "domain-analyst",
                             "next_step": "architect_site"
                         },
                         "architect_site": {
                             "action": "call_agent",
                             "description": "Design site architecture",
                             "agent_type": "site-architect",
                             "next_step": "gather_content"
                         },
                         "gather_content": {
                             "action": "fan_out",
                             "description": "Gather content and visuals in parallel",
                             "sub_tasks": [
                                 {"agent_type": "content-researcher", "step_name": "research"},
                                 {"agent_type": "visual-designer", "step_name": "visuals"}
                             ],
                             "next_step": "develop_site"
                         },
                         "develop_site": {
                             "action": "call_agent",
                             "description": "Generate all HTML pages",
                             "agent_type": "html-developer",
                             "next_step": "publish_site"
                         },
                         "publish_site": {
                             "action": "call_agent",
                             "description": "Publish to hosting",
                             "agent_type": "site-publisher",
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Complete website creation"
                         }
                     }
                 }
             }'::jsonb,
             '["orchestration", "website-creation", "project-management"]'::jsonb,
             '{
                 "process": "system.agent.website-builder.process",
                 "response": "system.responses.website-builder",
                 "error": "system.errors.website-builder",
                 "dlq": "dlq.website-builder"
             }'::jsonb
         ) ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    image_repository = EXCLUDED.image_repository,
    image_tag = EXCLUDED.image_tag,
    command = EXCLUDED.command,
    resources = EXCLUDED.resources,
    default_config = EXCLUDED.default_config,
    capabilities = EXCLUDED.capabilities,
    topics = EXCLUDED.topics,
    updated_at = NOW();

-- 8. Content Creator Agent (with FULL memory and style config preserved)
INSERT INTO agent_definitions (
    type, display_name, description, category,
    image_repository, image_tag, command,
    resources, default_config, capabilities, topics, env_vars
) VALUES (
             'content-creator',
             'Content Creator',
             'Advanced AI-powered content generation with memory and style adaptation',
             'data-driven',
             'docker.io/aqls/content-creator-agent',
             'v1.0.48',
             ARRAY['./content-creator-agent', '-config', 'configs/content-creator-agent.yaml'],
             '{
                 "requests": {"cpu": "200m", "memory": "512Mi"},
                 "limits": {"cpu": "1000m", "memory": "2Gi"}
             }'::jsonb,
             '{
                 "model": "claude-3-5-sonnet-20241022",
                 "temperature": 0.7,
                 "max_tokens": 2000,
                 "processing_mode": "task",
                 "memory_config": {
                     "enabled": true,
                     "auto_store": true,
                     "auto_store_threshold": 0.7,
                     "max_memories": 100,
                     "retrieval_count": 5,
                     "embedding_model": "text-embedding-ada-002",
                     "include_types": ["generated_content", "user_feedback", "style_preferences"]
                 },
                 "metrics_config": {
                     "enabled": true,
                     "fail_silently": true,
                     "detailed_errors": false,
                     "record_token_usage": true,
                     "record_latency": true,
                     "record_errors": true
                 },
                 "workflow": {
                     "start_step": "retrieve_context",
                     "steps": {
                         "retrieve_context": {
                             "action": "retrieve_memory",
                             "description": "Retrieve relevant memories and context",
                             "config": {
                                 "memory_types": ["style_preferences", "user_feedback"],
                                 "max_results": 5
                             },
                             "next_step": "generate_content"
                         },
                         "generate_content": {
                             "action": "ai_text_generate_anthropic",
                             "description": "Generate text content using Anthropic LLM with memory context",
                             "store_memory": true,
                             "next_step": "quality_check"
                         },
                         "quality_check": {
                             "action": "validate_output",
                             "description": "Check content quality and compliance",
                             "config": {
                                 "check_grammar": true,
                                 "check_tone": true,
                                 "check_length": true
                             },
                             "next_step": "store_memory"
                         },
                         "store_memory": {
                             "action": "store_memory",
                             "description": "Store generated content in memory",
                             "config": {
                                 "memory_type": "generated_content",
                                 "include_metadata": true
                             },
                             "next_step": "complete_workflow"
                         },
                         "complete_workflow": {
                             "action": "complete_workflow",
                             "description": "Mark workflow as complete and store results"
                         }
                     }
                 },
                 "supported_content_types": [
                     "blog_post", "product_description", "social_media",
                     "email", "landing_page", "press_release", "technical_doc",
                     "ad_copy", "newsletter", "case_study", "whitepaper"
                 ],
                 "style_options": ["informative", "persuasive", "casual", "professional", "creative", "technical", "conversational"],
                 "tone_options": ["friendly", "formal", "conversational", "authoritative", "enthusiastic", "empathetic", "neutral"],
                 "length_options": ["micro", "short", "medium", "long", "comprehensive"],
                 "platform_support": ["generic", "twitter", "linkedin", "facebook", "instagram", "medium", "substack"],
                 "seo_optimization": {
                     "enabled": true,
                     "keyword_density": 0.02,
                     "meta_description": true,
                     "heading_optimization": true
                 }
             }'::jsonb,
             '["content-generation", "writing", "copywriting", "seo", "memory-enabled", "style-adaptive"]'::jsonb,
             '{
                 "process": "system.agent.content-creator.process",
                 "response": "system.responses.content-creator",
                 "error": "system.errors.content-creator",
                 "dlq": "dlq.content-creator",
                 "priority_high": "tasks.high.content-creator",
                 "priority_normal": "tasks.normal.content-creator",
                 "priority_low": "tasks.low.content-creator"
             }'::jsonb,
             '[
                 {"name": "CONTENT_CREATOR_MODE", "value": "production"},
                 {"name": "ENABLE_METRICS", "value": "true"},
                 {"name": "MEMORY_ENABLED", "value": "true"}
             ]'::jsonb
         ) ON CONFLICT (type) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    image_repository = EXCLUDED.image_repository,
    image_tag = EXCLUDED.image_tag,
    command = EXCLUDED.command,
    resources = EXCLUDED.resources,
    default_config = EXCLUDED.default_config,
    capabilities = EXCLUDED.capabilities,
    topics = EXCLUDED.topics,
    env_vars = EXCLUDED.env_vars,
    updated_at = NOW();

-- ============================================================================
-- PART 2: SUPPORTING TABLES (Only those not in your existing files)
-- ============================================================================

-- Agent group members table (links agents to groups)
CREATE TABLE IF NOT EXISTS agent_group_members (
                                                   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_name VARCHAR(255) NOT NULL,
    agent_type VARCHAR(100) NOT NULL REFERENCES agent_definitions(type),
    role VARCHAR(100) NOT NULL,
    required BOOLEAN DEFAULT true,
    config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(group_name, agent_type)
    );

-- Workflow templates table
CREATE TABLE IF NOT EXISTS workflow_templates (
                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    category VARCHAR(100),
    workflow_definition JSONB NOT NULL,
    default_config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
    );

-- Agent capabilities table for capability matching
CREATE TABLE IF NOT EXISTS agent_capabilities (
                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_type VARCHAR(100) NOT NULL REFERENCES agent_definitions(type),
    capability VARCHAR(255) NOT NULL,
    strength DECIMAL(3,2) DEFAULT 1.0 CHECK (strength >= 0 AND strength <= 1),
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(agent_type, capability)
    );

-- Agent dependencies table
CREATE TABLE IF NOT EXISTS agent_dependencies (
                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_type VARCHAR(100) NOT NULL REFERENCES agent_definitions(type),
    depends_on VARCHAR(100) NOT NULL,
    dependency_type VARCHAR(50) NOT NULL CHECK (dependency_type IN ('data', 'optional', 'orchestration')),
    config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(agent_type, depends_on)
    );

-- Agent metrics configuration table
CREATE TABLE IF NOT EXISTS agent_metrics_config (
                                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_type VARCHAR(100) NOT NULL REFERENCES agent_definitions(type) UNIQUE,
    metrics_enabled BOOLEAN DEFAULT true,
    collection_interval INTEGER DEFAULT 60,
    retention_days INTEGER DEFAULT 30,
    metrics_config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
    );

-- Default agent configurations table
CREATE TABLE IF NOT EXISTS agent_default_configs (
                                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    config_name VARCHAR(255) NOT NULL UNIQUE,
    agent_type VARCHAR(100) NOT NULL REFERENCES agent_definitions(type),
    environment VARCHAR(50) DEFAULT 'production',
    config JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
    );

-- ============================================================================
-- PART 3: DATA POPULATION
-- ============================================================================

-- Insert agent group members
INSERT INTO agent_group_members (group_name, agent_type, role, required, config) VALUES
                                                                                     ('website-builder', 'website-builder', 'orchestrator', true, '{"priority": 1, "spawn_mode": "single"}'::jsonb),
                                                                                     ('website-builder', 'domain-analyst', 'domain_analyst', true, '{"priority": 2, "spawn_mode": "single", "timeout_seconds": 300}'::jsonb),
                                                                                     ('website-builder', 'site-architect', 'site_architect', true, '{"priority": 3, "spawn_mode": "single", "timeout_seconds": 300}'::jsonb),
                                                                                     ('website-builder', 'content-researcher', 'content_researcher', true, '{"priority": 4, "spawn_mode": "parallel", "timeout_seconds": 600}'::jsonb),
                                                                                     ('website-builder', 'content-creator', 'content_writer', true, '{"priority": 5, "spawn_mode": "parallel", "timeout_seconds": 900}'::jsonb),
                                                                                     ('website-builder', 'html-developer', 'html_developer', true, '{"priority": 6, "spawn_mode": "single", "timeout_seconds": 1200}'::jsonb),
                                                                                     ('website-builder', 'visual-designer', 'visual_designer', false, '{"priority": 4, "spawn_mode": "parallel", "timeout_seconds": 600}'::jsonb),
                                                                                     ('website-builder', 'site-publisher', 'site_publisher', true, '{"priority": 7, "spawn_mode": "single", "timeout_seconds": 300}'::jsonb)
    ON CONFLICT (group_name, agent_type) DO UPDATE SET
    role = EXCLUDED.role,
                                                required = EXCLUDED.required,
                                                config = EXCLUDED.config,
                                                updated_at = NOW();

-- Insert workflow template
INSERT INTO workflow_templates (name, description, category, workflow_definition, default_config) VALUES (
                                                                                                             'website-creation-workflow',
                                                                                                             'Complete workflow for building a website from scratch',
                                                                                                             'website-builder',
                                                                                                             '{
                                                                                                                 "version": "1.0",
                                                                                                                 "start_step": "init",
                                                                                                                 "steps": {
                                                                                                                     "init": {
                                                                                                                         "type": "initialize",
                                                                                                                         "description": "Initialize website creation workflow",
                                                                                                                         "next": "analyze_domain"
                                                                                                                     },
                                                                                                                     "analyze_domain": {
                                                                                                                         "type": "agent_task",
                                                                                                                         "agent": "domain-analyst",
                                                                                                                         "description": "Analyze business domain and requirements",
                                                                                                                         "outputs": ["business_type", "keywords", "target_audience"],
                                                                                                                         "next": "design_architecture"
                                                                                                                     },
                                                                                                                     "design_architecture": {
                                                                                                                         "type": "agent_task",
                                                                                                                         "agent": "site-architect",
                                                                                                                         "description": "Design site structure and navigation",
                                                                                                                         "inputs": ["business_type", "keywords"],
                                                                                                                         "outputs": ["site_structure", "pages", "navigation"],
                                                                                                                         "next": "parallel_content_visual"
                                                                                                                     },
                                                                                                                     "parallel_content_visual": {
                                                                                                                         "type": "parallel",
                                                                                                                         "description": "Research content and create visuals in parallel",
                                                                                                                         "branches": [
                                                                                                                             {
                                                                                                                                 "name": "content_branch",
                                                                                                                                 "steps": ["research_content", "write_content"]
                                                                                                                             },
                                                                                                                             {
                                                                                                                                 "name": "visual_branch",
                                                                                                                                 "steps": ["design_visuals"]
                                                                                                                             }
                                                                                                                         ],
                                                                                                                         "next": "develop_html"
                                                                                                                     },
                                                                                                                     "research_content": {
                                                                                                                         "type": "agent_task",
                                                                                                                         "agent": "content-researcher",
                                                                                                                         "description": "Research and gather content information",
                                                                                                                         "inputs": ["site_structure", "keywords", "target_audience"],
                                                                                                                         "outputs": ["research_data", "competitor_analysis", "content_topics"]
                                                                                                                     },
                                                                                                                     "write_content": {
                                                                                                                         "type": "agent_task",
                                                                                                                         "agent": "content-creator",
                                                                                                                         "description": "Write content for all pages",
                                                                                                                         "inputs": ["research_data", "pages", "keywords"],
                                                                                                                         "outputs": ["page_content", "seo_metadata"]
                                                                                                                     },
                                                                                                                     "design_visuals": {
                                                                                                                         "type": "agent_task",
                                                                                                                         "agent": "visual-designer",
                                                                                                                         "description": "Create visual assets and branding",
                                                                                                                         "inputs": ["business_type", "pages"],
                                                                                                                         "outputs": ["logo", "images", "color_scheme"],
                                                                                                                         "required": false
                                                                                                                     },
                                                                                                                     "develop_html": {
                                                                                                                         "type": "agent_task",
                                                                                                                         "agent": "html-developer",
                                                                                                                         "description": "Generate HTML/CSS/JS for all pages",
                                                                                                                         "inputs": ["pages", "page_content", "images", "color_scheme"],
                                                                                                                         "outputs": ["html_files", "css_files", "js_files"],
                                                                                                                         "next": "publish_site"
                                                                                                                     },
                                                                                                                     "publish_site": {
                                                                                                                         "type": "agent_task",
                                                                                                                         "agent": "site-publisher",
                                                                                                                         "description": "Deploy website to hosting",
                                                                                                                         "inputs": ["html_files", "css_files", "js_files", "images"],
                                                                                                                         "outputs": ["site_url", "deployment_status"],
                                                                                                                         "next": "complete"
                                                                                                                     },
                                                                                                                     "complete": {
                                                                                                                         "type": "complete",
                                                                                                                         "description": "Website creation completed"
                                                                                                                     }
                                                                                                                 }
                                                                                                             }'::jsonb,
                                                                                                             '{
                                                                                                                 "timeout_minutes": 60,
                                                                                                                 "max_retries": 2,
                                                                                                                 "parallel_execution": true,
                                                                                                                 "store_intermediate_results": true,
                                                                                                                 "notification_channels": ["email", "webhook"]
                                                                                                             }'::jsonb
                                                                                                         ) ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    workflow_definition = EXCLUDED.workflow_definition,
    default_config = EXCLUDED.default_config,
    updated_at = NOW();

-- Insert agent capabilities
INSERT INTO agent_capabilities (agent_type, capability, strength, metadata) VALUES
                                                                                ('domain-analyst', 'business-analysis', 0.95, '{"specialties": ["domain-research", "market-analysis"]}'::jsonb),
                                                                                ('domain-analyst', 'keyword-research', 0.90, '{"tools": ["seo-analysis", "trend-detection"]}'::jsonb),
                                                                                ('site-architect', 'information-architecture', 0.95, '{"specialties": ["navigation-design", "user-flow"]}'::jsonb),
                                                                                ('site-architect', 'ux-design', 0.85, '{"focus": ["structure", "hierarchy"]}'::jsonb),
                                                                                ('content-researcher', 'deep-research', 0.95, '{"tools": ["perplexity", "web-scraping"]}'::jsonb),
                                                                                ('content-researcher', 'competitor-analysis', 0.90, '{"methods": ["firecrawl", "content-extraction"]}'::jsonb),
                                                                                ('content-creator', 'content-writing', 0.95, '{"types": ["blog", "marketing", "technical"]}'::jsonb),
                                                                                ('content-creator', 'seo-optimization', 0.85, '{"features": ["keyword-density", "meta-tags"]}'::jsonb),
                                                                                ('html-developer', 'frontend-development', 0.95, '{"languages": ["html", "css", "javascript"]}'::jsonb),
                                                                                ('html-developer', 'responsive-design', 0.90, '{"frameworks": ["bootstrap", "tailwind"]}'::jsonb),
                                                                                ('visual-designer', 'graphic-design', 0.90, '{"tools": ["ai-generation", "image-optimization"]}'::jsonb),
                                                                                ('visual-designer', 'branding', 0.85, '{"capabilities": ["logo-design", "color-schemes"]}'::jsonb),
                                                                                ('site-publisher', 'deployment', 0.95, '{"platforms": ["s3", "cloudfront", "netlify"]}'::jsonb),
                                                                                ('site-publisher', 'hosting-management', 0.90, '{"features": ["ssl", "cdn", "domain-mapping"]}'::jsonb),
                                                                                ('website-builder', 'orchestration', 0.95, '{"capabilities": ["workflow-management", "agent-coordination"]}'::jsonb),
                                                                                ('website-builder', 'project-management', 0.90, '{"features": ["timeline-tracking", "resource-allocation"]}'::jsonb)
    ON CONFLICT (agent_type, capability) DO UPDATE SET
    strength = EXCLUDED.strength,
                                                metadata = EXCLUDED.metadata,
                                                updated_at = NOW();

-- Insert agent dependencies
INSERT INTO agent_dependencies (agent_type, depends_on, dependency_type, config) VALUES
                                                                                     ('site-architect', 'domain-analyst', 'data', '{"required_outputs": ["business_type", "keywords"]}'::jsonb),
                                                                                     ('content-researcher', 'site-architect', 'data', '{"required_outputs": ["pages", "site_structure"]}'::jsonb),
                                                                                     ('content-creator', 'content-researcher', 'data', '{"required_outputs": ["research_data", "content_topics"]}'::jsonb),
                                                                                     ('html-developer', 'content-creator', 'data', '{"required_outputs": ["page_content"]}'::jsonb),
                                                                                     ('html-developer', 'visual-designer', 'optional', '{"optional_outputs": ["images", "color_scheme"]}'::jsonb),
                                                                                     ('site-publisher', 'html-developer', 'data', '{"required_outputs": ["html_files", "css_files"]}'::jsonb),
                                                                                     ('website-builder', 'ALL', 'orchestration', '{"coordinate_all": true}'::jsonb)
    ON CONFLICT (agent_type, depends_on) DO UPDATE SET
    dependency_type = EXCLUDED.dependency_type,
                                                config = EXCLUDED.config,
                                                updated_at = NOW();

-- Insert metrics configuration
INSERT INTO agent_metrics_config (agent_type, metrics_enabled, collection_interval, retention_days, metrics_config) VALUES
                                                                                                                        ('domain-analyst', true, 60, 30, '{"track": ["latency", "success_rate", "token_usage"], "alert_threshold": 0.95}'::jsonb),
                                                                                                                        ('site-architect', true, 60, 30, '{"track": ["latency", "success_rate", "pages_designed"], "alert_threshold": 0.95}'::jsonb),
                                                                                                                        ('content-researcher', true, 60, 30, '{"track": ["latency", "api_calls", "data_gathered"], "alert_threshold": 0.90}'::jsonb),
                                                                                                                        ('content-creator', true, 60, 30, '{"track": ["latency", "token_usage", "content_quality"], "alert_threshold": 0.90}'::jsonb),
                                                                                                                        ('html-developer', true, 60, 30, '{"track": ["latency", "code_lines", "validation_errors"], "alert_threshold": 0.95}'::jsonb),
                                                                                                                        ('visual-designer', true, 60, 30, '{"track": ["latency", "images_processed", "api_calls"], "alert_threshold": 0.85}'::jsonb),
                                                                                                                        ('site-publisher', true, 60, 30, '{"track": ["latency", "upload_size", "deployment_time"], "alert_threshold": 0.99}'::jsonb),
                                                                                                                        ('website-builder', true, 30, 90, '{"track": ["total_duration", "agents_spawned", "success_rate"], "alert_threshold": 0.95}'::jsonb)
    ON CONFLICT (agent_type) DO UPDATE SET
    metrics_enabled = EXCLUDED.metrics_enabled,
                                    collection_interval = EXCLUDED.collection_interval,
                                    retention_days = EXCLUDED.retention_days,
                                    metrics_config = EXCLUDED.metrics_config,
                                    updated_at = NOW();

-- Insert default configurations
INSERT INTO agent_default_configs (config_name, agent_type, environment, config) VALUES
                                                                                     ('production-domain-analyst', 'domain-analyst', 'production', '{"model": "claude-3-5-sonnet-20241022", "temperature": 0.3, "max_retries": 3}'::jsonb),
                                                                                     ('production-site-architect', 'site-architect', 'production', '{"model": "claude-3-5-sonnet-20241022", "temperature": 0.5, "max_pages": 20}'::jsonb),
                                                                                     ('production-content-researcher', 'content-researcher', 'production', '{"search_depth": 3, "competitor_limit": 5, "include_sources": true}'::jsonb),
                                                                                     ('production-content-creator', 'content-creator', 'production', '{"model": "claude-3-5-sonnet-20241022", "temperature": 0.7, "enable_memory": true}'::jsonb),
                                                                                     ('production-html-developer', 'html-developer', 'production', '{"framework": "vanilla", "responsive": true, "minify": true}'::jsonb),
                                                                                     ('production-visual-designer', 'visual-designer', 'production', '{"image_quality": "high", "optimize": true, "formats": ["webp", "jpg"]}'::jsonb),
                                                                                     ('production-site-publisher', 'site-publisher', 'production', '{"cdn_enabled": true, "ssl": true, "compression": true}'::jsonb),
                                                                                     ('production-website-builder', 'website-builder', 'production', '{"max_duration_minutes": 60, "parallel_agents": 5, "notification": true}'::jsonb)
    ON CONFLICT (config_name) DO UPDATE SET
    config = EXCLUDED.config,
                                     updated_at = NOW();

-- ============================================================================
-- PART 4: INDEXES FOR PERFORMANCE
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_agent_group_members_group ON agent_group_members(group_name);
CREATE INDEX IF NOT EXISTS idx_agent_group_members_agent ON agent_group_members(agent_type);
CREATE INDEX IF NOT EXISTS idx_workflow_templates_name ON workflow_templates(name);
CREATE INDEX IF NOT EXISTS idx_workflow_templates_category ON workflow_templates(category);
CREATE INDEX IF NOT EXISTS idx_agent_capabilities_agent ON agent_capabilities(agent_type);
CREATE INDEX IF NOT EXISTS idx_agent_dependencies_agent ON agent_dependencies(agent_type);
CREATE INDEX IF NOT EXISTS idx_agent_metrics_config_agent ON agent_metrics_config(agent_type);
CREATE INDEX IF NOT EXISTS idx_agent_default_configs_name ON agent_default_configs(config_name);
CREATE INDEX IF NOT EXISTS idx_agent_default_configs_agent ON agent_default_configs(agent_type);

-- ============================================================================
-- PART 5: COMPLETION MESSAGE
-- ============================================================================

DO $$
BEGIN
    RAISE NOTICE 'Website Builder Agents migration completed successfully';
    RAISE NOTICE 'Total agents configured: 8';
    RAISE NOTICE 'Supporting tables created: 7';
    RAISE NOTICE 'Workflow template created: website-creation-workflow';
END $$;