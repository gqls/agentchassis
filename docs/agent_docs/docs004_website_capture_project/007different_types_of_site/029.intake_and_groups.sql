-- ============================================================================
-- INTAKE ORCHESTRATOR + SPAWN_GROUP ARCHITECTURE
-- ============================================================================

-- ============================================================================
-- 1. SCHEMA UPDATE: Add briefing_questionnaire to agent_group_definitions
-- ============================================================================

ALTER TABLE agent_group_definitions
    ADD COLUMN IF NOT EXISTS briefing_questionnaire JSONB DEFAULT '{}'::jsonb;

COMMENT ON COLUMN agent_group_definitions.briefing_questionnaire IS
'Questionnaire definition for the briefing agent to execute for this group type';

-- ============================================================================
-- 2. SITE CLASSIFIER AGENT
-- Lightweight agent that determines site_type from domain/objective
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config
)
VALUES (
           gen_random_uuid(),
           'site-classifier',
           'Site Classifier',
           'Analyzes domain and objective to determine site type and recommend appropriate builder group',
           'classification',
           '{
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
                     "input_fields": ["input_data"],
                     "output_field": "classification",
                     "prompt_template": "Classify this website project and recommend the appropriate builder.\n\nInput:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n\nClassify into ONE of these site types:\n\n**landing** - Conversion-focused single-purpose sites:\n- Product/service sales pages\n- SaaS landing pages, app downloads\n- Lead generation, signups\n- Event registration\n- Clear single CTA goal\n- Examples: stripe.com landing, mailchimp signup, product launches\n\n**content** - Publishing/content sites:\n- News, blogs, magazines, articles\n- Content aggregation\n- Ad-revenue or subscription models\n- SEO/traffic focused\n- Category navigation, archives\n- Examples: medium.com, techcrunch, recipe blogs\n\n**portfolio** - Showcase/portfolio sites:\n- Creative portfolios, agencies\n- Case studies, project galleries\n- Visual/image heavy\n- Client testimonials\n- Examples: dribbble profiles, agency sites\n\n**brochure** - Multi-page business sites:\n- Company websites with multiple pages\n- About, services, team, contact\n- Informational focus\n- Examples: law firms, consultancies, local businesses\n\nAnalyze the domain name and stated objective to determine the best fit.\n\nReturn ONLY valid JSON:\n{\n  \"site_type\": \"landing|content|portfolio|brochure\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"Brief explanation of classification\",\n  \"recommended_group\": \"landing-page-builder|content-site-builder|portfolio-builder|brochure-builder\",\n  \"detected_industry\": \"Industry/niche if detectable\",\n  \"detected_signals\": [\"Signal 1 from domain/objective\", \"Signal 2\"]\n}"
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
           }'::jsonb,
           true,
           '["classification", "analysis", "llm"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.476',
           '{"limits": {"cpu": "250m", "memory": "256Mi"}, "requests": {"cpu": "50m", "memory": "128Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              updated_at = now();

-- ============================================================================
-- 3. BRIEFING AGENT - Executes questionnaires via LLM or HITL
-- ============================================================================
-- updated later in doc 030.briefing_agent.sql

/*INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config
)
VALUES (
           gen_random_uuid(),
           'briefing-agent',
           'Briefing Agent',
           'Executes briefing questionnaires - either via LLM inference or HITL collection',
           'data-collection',
           '{
             "workflow": {
               "start_step": "check_mode",
               "steps": {
                 "check_mode": {
                   "action": "evaluate_condition",
                   "config": {
                     "condition_field": "input_data.hitl_mode",
                     "conditions": {
                       "interactive": "collect_via_hitl",
                       "auto": "infer_via_llm"
                     },
                     "default": "infer_via_llm"
                   },
                   "next_step": "infer_via_llm",
                   "description": "Determine if briefing should be interactive or auto-inferred"
                 },
                 "infer_via_llm": {
                   "action": "execute_llm_prompt",
                   "config": {
                     "ai_service": {
                       "provider": "anthropic",
                       "model": "claude-haiku-4-5-20251001",
                       "api_key_env_var": "ANTHROPIC_API_KEY",
                       "max_tokens": 4000
                     },
                     "input_fields": ["input_data", "questionnaire"],
                     "output_field": "brief_answers",
                     "prompt_template": "You are completing a briefing questionnaire for a website project.\n\nProject Info:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Site Type: {{.input_data.site_type}}\n- Industry: {{.input_data.detected_industry}}\n\nQuestionnaire to complete:\n{{.questionnaire}}\n\nBased on the domain name, objective, and your knowledge of the industry, provide thoughtful answers to each question in the questionnaire.\n\nFor questions you cannot confidently answer from the available information, provide reasonable defaults appropriate for the industry.\n\nReturn your answers as a JSON object where keys match the field names in the questionnaire:\n{\n  \"field_name_1\": \"your answer\",\n  \"field_name_2\": \"your answer\",\n  ...\n}\n\nReturn ONLY valid JSON."
                   },
                   "next_step": "complete"
                 },
                 "collect_via_hitl": {
                   "action": "request_human_input",
                   "config": {
                     "request_type": "questionnaire",
                     "questionnaire_field": "questionnaire",
                     "timeout_seconds": 86400,
                     "message": "Please complete the briefing questionnaire for this project"
                   },
                   "output_field": "brief_answers",
                   "next_step": "complete",
                   "description": "Collect answers via human-in-the-loop"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Return the completed brief"
                 }
               }
             },
             "processing_mode": "task",
             "timeout_seconds": 120
           }'::jsonb,
           true,
           '["briefing", "questionnaire", "llm", "hitl"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.476',
           '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              updated_at = now();*/

-- ============================================================================
-- 4. INTAKE ORCHESTRATOR GROUP
-- The entry point that classifies, briefs, and spawns the appropriate builder
-- ============================================================================

INSERT INTO agent_group_definitions (
    id, name, group_type, version, description,
    agent_configs, orchestration_workflow, briefing_questionnaire
)
VALUES (
           gen_random_uuid(),
           'Intake Orchestrator',
           'intake-orchestrator',
           1,
           'Entry point for site creation: classifies project type, runs briefing, spawns appropriate builder group',
           '[
             {"role": "classifier", "agent_type": "site-classifier"},
             {"role": "briefer", "agent_type": "briefing-agent"}
           ]'::jsonb,
           '{
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
                       "label": "Builder Group",
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
                 "action": "fetch_group_questionnaire",
                 "config": {
                   "group_type_field": "confirmed_type.recommended_group"
                 },
                 "output_field": "questionnaire",
                 "next_step": "call_briefer",
                 "description": "Fetch the briefing questionnaire for the target group"
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
                 "next_step": "spawn_builder_group",
                 "description": "Human reviews and can edit the brief before proceeding"
               },
               "spawn_builder_group": {
                 "action": "spawn_group",
                 "config": {
                   "group_type_field": "confirmed_type.recommended_group",
                   "input_fields": ["input_data", "classification", "brief_data", "reviewed_brief"],
                   "wait_for_completion": false
                 },
                 "output_field": "spawned_group",
                 "next_step": "complete",
                 "description": "Spawn the appropriate builder group with all collected data"
               },
               "complete": {
                 "action": "complete_workflow",
                 "description": "Intake complete - builder group has been spawned"
               }
             }
           }'::jsonb,
           '{}'::jsonb  -- Intake has no questionnaire - it fetches from target group
       )
    ON CONFLICT (group_type, version) DO UPDATE SET
    orchestration_workflow = EXCLUDED.orchestration_workflow,
                                    agent_configs = EXCLUDED.agent_configs,
                                    version = agent_group_definitions.version + 1,
                                    updated_at = now();

-- ============================================================================
-- 5. LANDING PAGE BUILDER GROUP (renamed from mvp-site-builder)
-- ============================================================================

INSERT INTO agent_group_definitions (
    id, name, group_type, version, description,
    agent_configs, orchestration_workflow, briefing_questionnaire
)
VALUES (
           gen_random_uuid(),
           'Landing Page Builder',
           'landing-page-builder',
           1,
           'Builds conversion-focused landing pages with clear CTAs',
           '[
             {"role": "strategist", "agent_type": "chief-strategist"},
             {"role": "architect", "agent_type": "landing-page-architect"},
             {"role": "writer", "agent_type": "content-writer"},
             {"role": "assembler", "agent_type": "html-assembler"},
             {"role": "deployer", "agent_type": "site-deployer"}
           ]'::jsonb,
           '{
             "start_step": "spawn_strategist",
             "steps": {
               "spawn_strategist": {
                 "action": "spawn_agent",
                 "config": {"role": "strategist", "agent_type": "chief-strategist"},
                 "next_step": "spawn_architect",
                 "description": "Spawn strategist"
               },
               "spawn_architect": {
                 "action": "spawn_agent",
                 "config": {"role": "architect", "agent_type": "landing-page-architect"},
                 "next_step": "spawn_writer",
                 "description": "Spawn landing page architect"
               },
               "spawn_writer": {
                 "action": "spawn_agent",
                 "config": {"role": "writer", "agent_type": "content-writer"},
                 "next_step": "spawn_assembler",
                 "description": "Spawn content writer"
               },
               "spawn_assembler": {
                 "action": "spawn_agent",
                 "config": {"role": "assembler", "agent_type": "html-assembler"},
                 "next_step": "spawn_deployer",
                 "description": "Spawn HTML assembler"
               },
               "spawn_deployer": {
                 "action": "spawn_agent",
                 "config": {"role": "deployer", "agent_type": "site-deployer"},
                 "next_step": "call_strategist",
                 "description": "Spawn deployer"
               },
               "call_strategist": {
                 "action": "call_agent",
                 "config": {
                   "agent_type": "chief-strategist",
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
                 "next_step": "call_deployer",
                 "description": "Assemble final HTML with CSS/JS"
               },
               "call_deployer": {
                 "action": "call_agent",
                 "config": {
                   "agent_type": "site-deployer",
                   "target_role": "deployer",
                   "input_fields": ["final_html", "input_data"],
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
           }'::jsonb,
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
                   {"field": "secondary_cta", "type": "text", "label": "Secondary CTA (e.g., Learn More)", "required": false},
                   {"field": "urgency_factor", "type": "text", "label": "Any urgency element? (Limited time, spots, etc.)", "required": false}
                 ]
               },
               {
                 "name": "social_proof",
                 "title": "Trust & Social Proof",
                 "questions": [
                   {"field": "has_testimonials", "type": "boolean", "label": "Do you have customer testimonials?", "default": false},
                   {"field": "client_count", "type": "text", "label": "Number of customers/users (if applicable)", "required": false},
                   {"field": "notable_clients", "type": "text", "label": "Notable clients or partners", "required": false},
                   {"field": "awards_certifications", "type": "text", "label": "Awards, certifications, or press mentions", "required": false}
                 ]
               },
               {
                 "name": "content",
                 "title": "Content Preferences",
                 "questions": [
                   {"field": "hero_style", "type": "select", "label": "Hero Section Style", "options": ["headline_focused", "image_focused", "video_background", "split_layout"], "default": "headline_focused"},
                   {"field": "include_pricing", "type": "boolean", "label": "Include pricing section?", "default": false},
                   {"field": "include_faq", "type": "boolean", "label": "Include FAQ section?", "default": true},
                   {"field": "include_features", "type": "boolean", "label": "Include features/benefits section?", "default": true}
                 ]
               }
             ]
           }'::jsonb
       )
    ON CONFLICT (group_type, version) DO UPDATE SET
    orchestration_workflow = EXCLUDED.orchestration_workflow,
                                    agent_configs = EXCLUDED.agent_configs,
                                    briefing_questionnaire = EXCLUDED.briefing_questionnaire,
                                    version = agent_group_definitions.version + 1,
                                    updated_at = now();


-- ============================================================================
-- 6. CONTENT SITE BUILDER GROUP
-- ============================================================================

INSERT INTO agent_group_definitions (
    id, name, group_type, version, description,
    agent_configs, orchestration_workflow, briefing_questionnaire
)
VALUES (
           gen_random_uuid(),
           'Content Site Builder',
           'content-site-builder',
           1,
           'Builds content/publishing sites with article grids, categories, and ad support',
           '[
             {"role": "strategist", "agent_type": "chief-strategist"},
             {"role": "architect", "agent_type": "content-site-architect"},
             {"role": "writer", "agent_type": "content-writer"},
             {"role": "assembler", "agent_type": "html-assembler"},
             {"role": "deployer", "agent_type": "site-deployer"}
           ]'::jsonb,
           '{
             "start_step": "spawn_strategist",
             "steps": {
               "spawn_strategist": {
                 "action": "spawn_agent",
                 "config": {"role": "strategist", "agent_type": "chief-strategist"},
                 "next_step": "spawn_architect",
                 "description": "Spawn strategist"
               },
               "spawn_architect": {
                 "action": "spawn_agent",
                 "config": {"role": "architect", "agent_type": "content-site-architect"},
                 "next_step": "spawn_writer",
                 "description": "Spawn content site architect"
               },
               "spawn_writer": {
                 "action": "spawn_agent",
                 "config": {"role": "writer", "agent_type": "content-writer"},
                 "next_step": "spawn_assembler",
                 "description": "Spawn content writer"
               },
               "spawn_assembler": {
                 "action": "spawn_agent",
                 "config": {"role": "assembler", "agent_type": "html-assembler"},
                 "next_step": "spawn_deployer",
                 "description": "Spawn HTML assembler"
               },
               "spawn_deployer": {
                 "action": "spawn_agent",
                 "config": {"role": "deployer", "agent_type": "site-deployer"},
                 "next_step": "call_strategist",
                 "description": "Spawn deployer"
               },
               "call_strategist": {
                 "action": "call_agent",
                 "config": {
                   "agent_type": "chief-strategist",
                   "target_role": "strategist",
                   "input_fields": ["input_data", "brief_data"],
                   "timeout_seconds": 120
                 },
                 "output_field": "build_plan",
                 "next_step": "call_architect",
                 "description": "Generate content site build plan"
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
                 "next_step": "call_writer",
                 "description": "Assemble content site template"
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
                 "description": "Generate content for template"
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
                 "next_step": "call_deployer",
                 "description": "Assemble final HTML"
               },
               "call_deployer": {
                 "action": "call_agent",
                 "config": {
                   "agent_type": "site-deployer",
                   "target_role": "deployer",
                   "input_fields": ["final_html", "input_data"],
                   "timeout_seconds": 180
                 },
                 "output_field": "deployment_result",
                 "next_step": "complete",
                 "description": "Deploy to git repository"
               },
               "complete": {
                 "action": "complete_workflow",
                 "description": "Content site build complete"
               }
             }
           }'::jsonb,
           '{
             "sections": [
               {
                 "name": "brand",
                 "title": "Publication Identity",
                 "questions": [
                   {"field": "publication_name", "type": "text", "label": "Publication/Site Name", "required": true},
                   {"field": "tagline", "type": "text", "label": "Tagline", "required": false},
                   {"field": "editorial_tone", "type": "select", "label": "Editorial Tone", "options": ["news_formal", "magazine_polished", "blog_casual", "technical", "opinion"], "default": "magazine_polished"}
                 ]
               },
               {
                 "name": "content_structure",
                 "title": "Content Structure",
                 "questions": [
                   {"field": "categories", "type": "textarea", "label": "Content Categories (one per line)", "required": true, "placeholder": "Technology\nBusiness\nLifestyle"},
                   {"field": "content_types", "type": "multiselect", "label": "Types of Content", "options": ["articles", "news", "opinion", "reviews", "guides", "lists"], "default": ["articles"]},
                   {"field": "publishing_frequency", "type": "select", "label": "Expected Publishing Frequency", "options": ["daily", "several_per_week", "weekly", "occasional"], "default": "weekly"}
                 ]
               },
               {
                 "name": "monetization",
                 "title": "Monetization",
                 "questions": [
                   {"field": "monetization_model", "type": "select", "label": "Primary Revenue Model", "options": ["advertising", "subscription", "affiliate", "sponsored", "none"], "default": "advertising"},
                   {"field": "ad_placements", "type": "multiselect", "label": "Ad Placement Zones", "options": ["sidebar", "in_content", "header_banner", "footer"], "default": ["sidebar"]},
                   {"field": "newsletter_signup", "type": "boolean", "label": "Include Newsletter Signup?", "default": true}
                 ]
               },
               {
                 "name": "features",
                 "title": "Site Features",
                 "questions": [
                   {"field": "include_search", "type": "boolean", "label": "Include Search?", "default": true},
                   {"field": "include_author_pages", "type": "boolean", "label": "Include Author Pages?", "default": false},
                   {"field": "include_comments", "type": "boolean", "label": "Include Comments? (future)", "default": false},
                   {"field": "include_related_articles", "type": "boolean", "label": "Show Related Articles?", "default": true},
                   {"field": "include_popular_sidebar", "type": "boolean", "label": "Show Popular/Trending Sidebar?", "default": true}
                 ]
               },
               {
                 "name": "sample_content",
                 "title": "Initial Content",
                 "questions": [
                   {"field": "generate_sample_articles", "type": "boolean", "label": "Generate Sample Articles?", "default": true},
                   {"field": "sample_article_count", "type": "number", "label": "Number of Sample Articles", "default": 6, "min": 0, "max": 20},
                   {"field": "featured_article_topic", "type": "text", "label": "Topic for Featured Article", "required": false}
                 ]
               }
             ]
           }'::jsonb
       )
    ON CONFLICT (group_type, version) DO UPDATE SET
    orchestration_workflow = EXCLUDED.orchestration_workflow,
                                    agent_configs = EXCLUDED.agent_configs,
                                    briefing_questionnaire = EXCLUDED.briefing_questionnaire,
                                    version = agent_group_definitions.version + 1,
                                    updated_at = now();


-- ============================================================================
-- 7. HELPER ACTION: fetch_group_questionnaire
-- Fetches the briefing questionnaire from a group definition
-- ============================================================================
-- This needs to be implemented as a Go action. See fetch_group_questionnaire.go

-- ============================================================================
-- 8. CLEAN UP OLD MVP-SITE-BUILDER (optional - keep for reference)
-- ============================================================================
-- UPDATE agent_group_definitions
-- SET group_type = 'mvp-site-builder-deprecated'
-- WHERE group_type = 'mvp-site-builder';