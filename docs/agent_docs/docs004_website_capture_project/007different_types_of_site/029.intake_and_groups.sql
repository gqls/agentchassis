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

INSERT INTO agent_definitions (
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
    ON CONFLICT (type,version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              updated_at = now();

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
             {"role": "strategist", "agent_type": "site-strategist"},
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
                 "config": {"role": "strategist", "agent_type": "site-strategist"},
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
             {"role": "strategist", "agent_type": "site-strategist"},
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
                 "config": {"role": "strategist", "agent_type": "site-strategist"},
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
                   "agent_type": "site-strategist",
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


-- ============================================================================
-- BRIEFING AGENT - INSERT (new agent definition)
-- Executes briefing questionnaires via LLM inference or HITL collection
-- ============================================================================

INSERT INTO agent_definitions (
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
                                       updated_at = now();


===

-- ============================================================================
-- NEW AGENT DEFINITIONS FOR INTAKE ORCHESTRATOR WORKFLOW
-- These are separate from the existing content-creator and deployer-agent
-- to avoid breaking the mvp-site-builder flow
-- ============================================================================

-- ============================================================================
-- 1. CONTENT WRITER - Works with briefing data and questionnaire answers
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config
)
VALUES (
           gen_random_uuid(),
           'content-writer',
           'Content Writer',
           'Creates website content from brief data and template requirements. Works with intake orchestrator flow.',
           'content',
           '{
             "workflow": {
               "start_step": "generate_content",
               "steps": {
                 "generate_content": {
                   "action": "execute_llm_prompt",
                   "config": {
                     "ai_service": {
                       "provider": "anthropic",
                       "model": "claude-haiku-4-5-20251001",
                       "api_key_env_var": "ANTHROPIC_API_KEY",
                       "max_tokens": 8000
                     },
                     "input_fields": ["template_data", "build_plan", "brief_data", "input_data"],
                     "output_field": "content_json",
                     "prompt_template": "You are a professional website content creator. Your job is to create compelling, industry-specific CONTENT (not HTML structure).\n\nWebsite Details:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n\nBrief Data (from questionnaire):\n{{.brief_data}}\n\nBuild Strategy (from strategist):\n{{.build_plan}}\n\nContent Requirements - these are the placeholders you need to fill:\n{{.template_data.content_requirements}}\n\nYour Task:\nCreate a JSON object with content for each placeholder. Group by component.\n\nGuidelines:\n- Write compelling, conversion-focused copy\n- Use the brief data to inform tone, messaging, and specifics\n- Match the domain and industry tone\n- For testimonials, use optimistic placeholder attributions like \"[Future You]\", \"[Soon-to-be Satisfied Customer]\" - NOT fake names\n- Use action-oriented language for CTAs\n- Keep brand consistency throughout\n- Stats/numbers should be truthful at all times, if there are no real stats available then do not use made up ones - we will have created maybe over 1000 of these sites.\"\n- If brief contains brand_name, use it consistently\n- If brief contains primary_cta, use that exact text for main call-to-action\n\nReturn ONLY valid JSON in this exact structure:\n{\n  \"meta\": {\n    \"title\": \"Page title for browser tab\",\n    \"description\": \"SEO meta description (150-160 chars)\"\n  },\n  \"theme\": \"recommended theme name from: default, clean-minimal, bold-conversion, warm-friendly, tech-saas, luxury-premium\",\n  \"theme_tags\": [\"semantic\", \"tags\", \"for\", \"theme\", \"matching\"],\n  \"sections\": {\n    \"component_header_0\": {\n      \"brand_name\": \"Your Brand Name\",\n      \"cta_text\": \"CTA Button Text\"\n    },\n    \"component_hero_1\": {\n      \"headline\": \"Main headline\",\n      \"subheadline\": \"Supporting text\",\n      \"primary_cta\": \"Primary button\",\n      \"secondary_cta\": \"Secondary button\"\n    }\n  }\n}\n\nFill ALL placeholders from the content requirements. Return ONLY the JSON object, no markdown or explanation."
                   },
                   "next_step": "complete"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Return the content JSON"
                 }
               }
             },
             "processing_mode": "task",
             "timeout_seconds": 300
           }'::jsonb,
           true,
           '["content", "copywriting", "llm"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.476',
           '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       updated_at = now();

-- ============================================================================
-- 2. SITE DEPLOYER - Works with new assembled HTML structure
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config
)
VALUES (
           gen_random_uuid(),
           'site-deployer',
           'Site Deployer',
           'Commits assembled site to git repository. Works with intake orchestrator flow.',
           'deployment',
           '{
             "workflow": {
               "start_step": "commit_to_git",
               "steps": {
                 "commit_to_git": {
                   "action": "git_commit",
                   "config": {
                     "repo_name": "sites",
                     "domain_field": "input_data.domain",
                     "input_fields": ["final_html", "input_data"],
                     "content_field": "final_html.html",
                     "commit_message": "Update site: {{.input_data.domain}}",
                     "filename": "index.html"
                   },
                   "next_step": "complete",
                   "description": "Commit to sites repo"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Deployment complete"
                 }
               }
             },
             "processing_mode": "task",
             "timeout_seconds": 180
           }'::jsonb,
           true,
           '["deployment", "git"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.476',
           '{"limits": {"cpu": "250m", "memory": "256Mi"}, "requests": {"cpu": "50m", "memory": "128Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       updated_at = now();


-- ============================================================================
-- 2b. SITE STRATEGIST - Works with briefing data to create enhanced build plans
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config
)
VALUES (
           gen_random_uuid(),
           'site-strategist',
           'Site Strategist',
           'Creates strategic build plans using behavioral psychology and briefing data. Works with intake orchestrator flow.',
           'strategy',
           '{
             "workflow": {
               "start_step": "generate_build_plan",
               "steps": {
                 "generate_build_plan": {
                   "action": "execute_llm_prompt",
                   "config": {
                     "ai_service": {
                       "provider": "anthropic",
                       "model": "claude-haiku-4-5-20251001",
                       "api_key_env_var": "ANTHROPIC_API_KEY",
                       "max_tokens": 3000
                     },
                     "input_fields": ["input_data", "brief_data"],
                     "output_field": "build_plan_json",
                     "prompt_template": "You are a website strategist creating a Build Plan based on behavioral psychology and conversion optimization.\n\nWebsite Request:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Behavioral Model: {{.input_data.model}}\n\n{{if .brief_data}}Detailed Brief (from analysis):\n{{.brief_data}}\n{{end}}\n\nYour Task:\nCreate a strategic Build Plan that maps behavioral psychology to website sections.\n\nBehavioral Models Available:\n- PAS (Problem-Agitate-Solution): Best for pain-point focused products\n- AIDA (Attention-Interest-Desire-Action): Best for general conversion\n- FAB (Features-Advantages-Benefits): Best for feature-rich products\n- 4Ps (Promise-Picture-Proof-Push): Best for aspirational products\n\nFor each model, define the appropriate sections:\n\nPAS sections: [\"problem_statement\", \"agitation\", \"solution_provider\", \"social_proof\", \"cta\"]\nAIDA sections: [\"attention_hero\", \"interest_features\", \"desire_benefits\", \"action_cta\"]\nFAB sections: [\"features_showcase\", \"advantages_comparison\", \"benefits_outcome\", \"cta\"]\n4Ps sections: [\"promise_hero\", \"picture_vision\", \"proof_testimonials\", \"push_cta\"]\n\n{{if .brief_data}}Use the briefing analysis to enhance your section selection:\n- Recommended theme: Use the theme.recommended from the brief\n- Messaging: Incorporate key_messages and usps into section guidance\n- Audience: Consider the target audience when defining section priorities\n{{end}}\n\nReturn ONLY valid JSON with this structure:\n{\n  \"model\": \"PAS|AIDA|FAB|4Ps\",\n  \"sections\": [\"section1\", \"section2\", ...],\n  \"section_guidance\": {\n    \"section_name\": {\n      \"purpose\": \"what this section should achieve\",\n      \"key_message\": \"primary message for this section\",\n      \"tone\": \"section-specific tone guidance\"\n    }\n  },\n  \"theme_recommendation\": \"recommended theme name\",\n  \"theme_tags\": [\"semantic\", \"tags\"],\n  \"conversion_priority\": [\"most important sections for conversion\"]\n}\n\nDO NOT include any text outside the JSON object."
                   },
                   "next_step": "complete"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Return the Build Plan"
                 }
               }
             },
             "processing_mode": "task",
             "timeout_seconds": 120
           }'::jsonb,
           true,
           '["strategy", "planning", "llm"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.476',
           '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              updated_at = now();


-- ============================================================================
-- 3. HTML ASSEMBLER - Assembles final HTML with CSS/JS injection
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config
)
VALUES (
           gen_random_uuid(),
           'html-assembler',
           'HTML Assembler',
           'Assembles final HTML from template and content, injects CSS themes and JS snippets.',
           'assembly',
           '{
             "workflow": {
               "start_step": "assemble_html",
               "steps": {
                 "assemble_html": {
                   "action": "assemble_full_page",
                   "config": {
                     "template_field": "template_data.html_template",
                     "content_field": "content_data.content_json",
                     "theme_field": "content_data.content_json.result.theme",
                     "theme_tags_field": "content_data.content_json.result.theme_tags",
                     "inject_css": true,
                     "inject_js": true,
                     "minify": false
                   },
                   "output_field": "final_html",
                   "next_step": "complete",
                   "description": "Assemble HTML with CSS and JS"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Return assembled HTML"
                 }
               }
             },
             "processing_mode": "task",
             "timeout_seconds": 120
           }'::jsonb,
           true,
           '["html", "assembly", "css", "javascript"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.476',
           '{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                              description = EXCLUDED.description,
                              updated_at = now();




====






====


-- ============================================================================
-- SPECIALIST ARCHITECT SYSTEM
-- ============================================================================

-- ============================================================================
-- 1. UPDATE BRIEFING AGENT - Now detects site_type
-- ============================================================================

/*UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = '{
    "workflow": {
      "start_step": "analyze_and_brief",
      "steps": {
        "analyze_and_brief": {
          "action": "execute_llm_prompt",
          "config": {
            "ai_service": {
              "provider": "anthropic",
              "model": "claude-haiku-4-5-20251001",
              "api_key_env_var": "ANTHROPIC_API_KEY",
              "max_tokens": 3000
            },
            "input_fields": ["input_data"],
            "output_field": "structured_brief",
            "prompt_template": "Analyze this website request and create a comprehensive brief.\n\nInput:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Model (if specified): {{.input_data.model}}\n\nFirst, determine the SITE TYPE based on the domain and objective:\n\n**landing** - Conversion-focused sites:\n- Product/service sales pages\n- SaaS landing pages\n- Lead generation\n- App download pages\n- Event registration\n- Single clear CTA goal\n\n**content** - Content/publishing sites:\n- News, blogs, magazines\n- Content aggregation\n- Ad-revenue models\n- SEO/traffic focused\n- Multiple articles/posts\n- Category navigation\n\n**portfolio** - Showcase sites:\n- Creative portfolios\n- Agency showcases\n- Case study focused\n- Gallery/visual heavy\n- Client testimonials\n\n**directory** - Listing sites:\n- Business directories\n- Job boards\n- Marketplace listings\n- Search/filter focused\n\nAnalyze the domain name and objective to determine site_type, then create a full brief.\n\nReturn JSON:\n{\n  \"site_type\": \"landing|content|portfolio|directory\",\n  \"site_type_confidence\": 0.0-1.0,\n  \"site_type_reasoning\": \"why this classification\",\n  \"analysis\": {\n    \"industry\": \"detected industry/niche\",\n    \"industry_confidence\": 0.0-1.0,\n    \"domain_interpretation\": \"what the domain name suggests\"\n  },\n  \"audience\": {\n    \"primary\": \"primary target audience\",\n    \"secondary\": \"secondary audience if applicable\",\n    \"demographics\": [\"age range\", \"profession\"],\n    \"psychographics\": [\"values\", \"motivations\", \"pain points\"]\n  },\n  \"brand\": {\n    \"tone\": \"professional|casual|technical|friendly|authoritative|playful\",\n    \"personality\": [\"trait1\", \"trait2\", \"trait3\"],\n    \"voice_examples\": [\"example phrase in brand voice\"]\n  },\n  \"messaging\": {\n    \"value_proposition\": \"core value proposition\",\n    \"key_messages\": [\"message1\", \"message2\", \"message3\"],\n    \"usps\": [\"unique selling point 1\", \"usp 2\"],\n    \"proof_points\": [\"credibility element 1\", \"element 2\"]\n  },\n  \"structure\": {\n    \"recommended_sections\": [\"section1\", \"section2\"],\n    \"priority_sections\": [\"most important sections\"],\n    \"optional_sections\": [\"sections that could be added\"]\n  },\n  \"theme\": {\n    \"recommended\": \"theme name\",\n    \"semantic_tags\": [\"tag1\", \"tag2\", \"tag3\"],\n    \"color_mood\": \"color feeling description\",\n    \"style_notes\": \"specific style recommendations\"\n  },\n  \"content_guidelines\": {\n    \"headline_style\": \"guidance for headlines\",\n    \"cta_style\": \"guidance for calls to action\",\n    \"avoid\": [\"things to avoid\"],\n    \"emphasize\": [\"things to emphasize\"]\n  },\n  \"monetization\": {\n    \"model\": \"subscription|ads|sales|leads|freemium\",\n    \"ad_zones\": [\"if ads: recommended ad placements\"]\n  }\n}\n\nReturn ONLY valid JSON."
          },
          "next_step": "complete"
        },
        "complete": {
          "action": "complete_workflow",
          "description": "Return the structured brief with site_type"
        }
      }
    },
    "processing_mode": "task",
    "timeout_seconds": 60
  }'::jsonb
WHERE type = 'briefing-agent';*/

-- ============================================================================
-- 2. RENAME CURRENT ARCHITECT → LANDING-PAGE-ARCHITECT
-- ============================================================================

-- First, create the new landing-page-architect as a copy
INSERT INTO agent_definitions (
    id,
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config
)
VALUES (
           gen_random_uuid(),
           'landing-page-architect',
           'Landing Page Architect',
           'Assembles conversion-focused landing pages from component library (PAS, AIDA, etc.)',

           -- FROM site-component-architect
           'data-driven',   -- category

           -- default_config copied exactly
           '{
             "workflow": {
               "steps": {
                 "complete": {
                   "action": "complete_workflow"
                 },
                 "assemble_template": {
                   "action": "assemble_from_library",
                   "config": {
                     "input_fields": ["build_plan_data"]
                   },
                   "next_step": "complete"
                 }
               },
               "start_step": "assemble_template"
             },
             "processing_mode": "task",
             "timeout_seconds": 120
           }'::jsonb,

           true,

           -- NEW capabilities for landing-page-architect
           '["build", "assemble", "database", "landing-page", "conversion"]'::jsonb,

           -- These copied from site-component-architect
           'docker.io/aqls/agent-chassis',
           'v1.0.478',
           '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       capabilities = EXCLUDED.capabilities,
                                       updated_at = now();


-- ============================================================================
-- 3. CREATE CONTENT-SITE-ARCHITECT
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config
)
VALUES (
           gen_random_uuid(),
           'content-site-architect',
           'Content Site Architect',
           'Assembles content/publishing sites with article grids, category nav, and ad zones',
           'data-driven',
           '{
             "workflow": {
               "start_step": "assemble_content_site",
               "steps": {
                 "assemble_content_site": {
                   "action": "assemble_from_library",
                   "config": {
                     "site_type": "content",
                     "input_fields": ["build_plan_data", "brief_data"],
                     "default_sections": ["header", "featured_article", "article_grid", "sidebar", "category_nav", "footer"],
                     "component_category": "content-site"
                   },
                   "next_step": "complete",
                   "description": "Assemble content site template from component library"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Return the content site template"
                 }
               }
             },
             "processing_mode": "task",
             "timeout_seconds": 180
           }'::jsonb,
           true,
           '["build", "assemble", "database", "content-site", "publishing"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'v1.0.476',
           '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       updated_at = now();


-- ============================================================================
-- 4. CREATE PORTFOLIO-ARCHITECT (for future use)
-- ============================================================================

INSERT INTO agent_definitions (
    id, type, display_name, description, category,
    default_config, is_active, capabilities,
    image_repository, image_tag, resources, topics, health_config
)
VALUES (
           gen_random_uuid(),
           'portfolio-architect',
           'Portfolio Site Architect',
           'Assembles portfolio/showcase sites with galleries, case studies, and visual layouts',
           'data-driven',
           '{
             "workflow": {
               "start_step": "assemble_portfolio_site",
               "steps": {
                 "assemble_portfolio_site": {
                   "action": "assemble_from_library",
                   "config": {
                     "site_type": "portfolio",
                     "input_fields": ["build_plan_data", "brief_data"],
                     "default_sections": ["header", "hero_visual", "work_grid", "case_study", "client_logos", "about", "contact", "footer"],
                     "component_category": "portfolio-site"
                   },
                   "next_step": "complete",
                   "description": "Assemble portfolio site template from component library"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "description": "Return the portfolio site template"
                 }
               }
             },
             "processing_mode": "task",
             "timeout_seconds": 180
           }'::jsonb,
           true,
           '["build", "assemble", "database", "portfolio-site", "showcase"]',
           'docker.io/aqls/agent-chassis',
           'v1.0.476',
           '{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
           '{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
           '{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb
       )
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       updated_at = now();

-- ============================================================================
-- 5. CONTENT SITE COMPONENTS
-- ============================================================================

-- Add category for content site components
INSERT INTO content_components (name, display_name, function, category, semantic_tags, html_template, input_schema, sort_order)
VALUES
-- Header for content sites (with category nav)
(
    'content_header',
    'Content Site Header',
    'header',
    'content-site',
    '["content", "publishing", "news"]'::jsonb,
    '<header class="site-header site-header--content">
      <div class="container">
        <nav class="site-header__nav">
          <a href="/" class="site-header__brand">{{.brand_name}}</a>
          <ul class="site-header__categories">
            {{range .categories}}
            <li><a href="#{{.slug}}" class="site-header__category">{{.name}}</a></li>
            {{end}}
          </ul>
          <div class="site-header__actions">
            <button class="site-header__search-toggle" aria-label="Search">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.35-4.35"/></svg>
            </button>
            {{if .show_subscribe}}
            <a href="#subscribe" class="button button--small button--primary">{{.subscribe_text}}</a>
            {{end}}
          </div>
        </nav>
      </div>
    </header>',
    '{"brand_name": "string", "categories": "array", "show_subscribe": "boolean", "subscribe_text": "string"}',
    10
),
-- Featured Article Hero
(
    'featured_article',
    'Featured Article Hero',
    'featured_content',
    'content-site',
    '["content", "hero", "featured"]'::jsonb,
    '<section class="section section--featured">
      <div class="container">
        <article class="featured-article">
          <div class="featured-article__image">
            <img src="{{.featured_image}}" alt="{{.featured_title}}" loading="lazy">
            <span class="featured-article__category">{{.featured_category}}</span>
          </div>
          <div class="featured-article__content">
            <h1 class="featured-article__title">{{.featured_title}}</h1>
            <p class="featured-article__excerpt">{{.featured_excerpt}}</p>
            <div class="featured-article__meta">
              <span class="featured-article__author">{{.featured_author}}</span>
              <span class="featured-article__date">{{.featured_date}}</span>
              <span class="featured-article__read-time">{{.featured_read_time}}</span>
            </div>
            <a href="#" class="button button--primary">{{.read_more_text}}</a>
          </div>
        </article>
      </div>
    </section>',
    '{"featured_image": "string", "featured_title": "string", "featured_excerpt": "string", "featured_category": "string", "featured_author": "string", "featured_date": "string", "featured_read_time": "string", "read_more_text": "string"}',
    20
),
-- Article Grid
(
    'article_grid',
    'Article Grid',
    'content_listing',
    'content-site',
    '["content", "grid", "articles"]'::jsonb,
    '<section class="section section--articles">
      <div class="container">
        <div class="section__header">
          <h2 class="section__title">{{.section_title}}</h2>
          {{if .section_subtitle}}<p class="section__subtitle">{{.section_subtitle}}</p>{{end}}
        </div>
        <div class="article-grid grid grid--3">
          {{range .articles}}
          <article class="article-card hover-lift">
            <div class="article-card__image">
              <img src="{{.image}}" alt="{{.title}}" loading="lazy">
              <span class="article-card__category">{{.category}}</span>
            </div>
            <div class="article-card__content">
              <h3 class="article-card__title">{{.title}}</h3>
              <p class="article-card__excerpt">{{.excerpt}}</p>
              <div class="article-card__meta">
                <span class="article-card__date">{{.date}}</span>
                <span class="article-card__read-time">{{.read_time}}</span>
              </div>
            </div>
          </article>
          {{end}}
        </div>
        {{if .show_load_more}}
        <div class="section__actions">
          <button class="button button--secondary">{{.load_more_text}}</button>
        </div>
        {{end}}
      </div>
    </section>',
    '{"section_title": "string", "section_subtitle": "string", "articles": "array", "show_load_more": "boolean", "load_more_text": "string"}',
    30
),
-- Sidebar with Ad Zone
(
    'content_sidebar',
    'Content Sidebar',
    'sidebar',
    'content-site',
    '["content", "sidebar", "ads"]'::jsonb,
    '<aside class="sidebar">
      {{if .show_newsletter}}
      <div class="sidebar__widget sidebar__widget--newsletter">
        <h3 class="sidebar__title">{{.newsletter_title}}</h3>
        <p class="sidebar__text">{{.newsletter_description}}</p>
        <form class="newsletter-form">
          <input type="email" placeholder="{{.email_placeholder}}" class="newsletter-form__input" required>
          <button type="submit" class="button button--primary button--full-width">{{.subscribe_button}}</button>
        </form>
      </div>
      {{end}}

      {{if .show_popular}}
      <div class="sidebar__widget sidebar__widget--popular">
        <h3 class="sidebar__title">{{.popular_title}}</h3>
        <ul class="popular-list">
          {{range .popular_articles}}
          <li class="popular-list__item">
            <a href="#" class="popular-list__link">
              <span class="popular-list__number">{{.rank}}</span>
              <span class="popular-list__title">{{.title}}</span>
            </a>
          </li>
          {{end}}
        </ul>
      </div>
      {{end}}

      {{if .show_ad}}
      <div class="sidebar__widget sidebar__widget--ad">
        <div class="ad-zone ad-zone--sidebar" data-ad-slot="{{.ad_slot_id}}">
          <span class="ad-zone__label">Advertisement</span>
          <!-- Ad content inserted here -->
        </div>
      </div>
      {{end}}

      {{if .show_categories}}
      <div class="sidebar__widget sidebar__widget--categories">
        <h3 class="sidebar__title">{{.categories_title}}</h3>
        <ul class="category-list">
          {{range .category_links}}
          <li><a href="#{{.slug}}" class="category-list__link">{{.name}} <span class="category-list__count">({{.count}})</span></a></li>
          {{end}}
        </ul>
      </div>
      {{end}}
    </aside>',
    '{"show_newsletter": "boolean", "newsletter_title": "string", "newsletter_description": "string", "email_placeholder": "string", "subscribe_button": "string", "show_popular": "boolean", "popular_title": "string", "popular_articles": "array", "show_ad": "boolean", "ad_slot_id": "string", "show_categories": "boolean", "categories_title": "string", "category_links": "array"}',
    40
),
-- In-content Ad Zone
(
    'ad_zone_inline',
    'Inline Ad Zone',
    'advertising',
    'content-site',
    '["ads", "monetization"]'::jsonb,
    '<div class="ad-zone ad-zone--inline" data-ad-slot="{{.ad_slot_id}}">
      <span class="ad-zone__label">Advertisement</span>
      <!-- Ad content inserted here -->
    </div>',
    '{"ad_slot_id": "string"}',
    50
),
-- Category Section
(
    'category_section',
    'Category Section',
    'category_listing',
    'content-site',
    '["content", "category", "navigation"]'::jsonb,
    '<section class="section section--category" id="{{.category_slug}}">
      <div class="container">
        <div class="section__header section__header--with-link">
          <h2 class="section__title">{{.category_name}}</h2>
          <a href="#" class="section__link">View all {{.category_name}} →</a>
        </div>
        <div class="article-grid grid grid--4">
          {{range .category_articles}}
          <article class="article-card article-card--compact hover-lift">
            <div class="article-card__image">
              <img src="{{.image}}" alt="{{.title}}" loading="lazy">
            </div>
            <div class="article-card__content">
              <h3 class="article-card__title">{{.title}}</h3>
              <span class="article-card__date">{{.date}}</span>
            </div>
          </article>
          {{end}}
        </div>
      </div>
    </section>',
    '{"category_slug": "string", "category_name": "string", "category_articles": "array"}',
    60
),
-- Content Site Footer
(
    'content_footer',
    'Content Site Footer',
    'footer',
    'content-site',
    '["content", "footer"]'::jsonb,
    '<footer class="site-footer site-footer--content">
      <div class="container">
        <div class="site-footer__grid grid grid--4">
          <div class="site-footer__about">
            <h3 class="site-footer__brand">{{.brand_name}}</h3>
            <p class="site-footer__tagline">{{.tagline}}</p>
            <div class="site-footer__social">
              {{range .social_links}}
              <a href="{{.url}}" class="site-footer__social-link" aria-label="{{.platform}}">{{.icon}}</a>
              {{end}}
            </div>
          </div>
          <div class="site-footer__links">
            <h4 class="site-footer__heading">Categories</h4>
            <ul class="site-footer__list">
              {{range .categories}}
              <li><a href="#{{.slug}}" class="site-footer__link">{{.name}}</a></li>
              {{end}}
            </ul>
          </div>
          <div class="site-footer__links">
            <h4 class="site-footer__heading">Company</h4>
            <ul class="site-footer__list">
              {{range .company_links}}
              <li><a href="{{.url}}" class="site-footer__link">{{.name}}</a></li>
              {{end}}
            </ul>
          </div>
          <div class="site-footer__newsletter">
            <h4 class="site-footer__heading">{{.newsletter_title}}</h4>
            <p class="site-footer__text">{{.newsletter_description}}</p>
            <form class="newsletter-form newsletter-form--footer">
              <input type="email" placeholder="{{.email_placeholder}}" class="newsletter-form__input">
              <button type="submit" class="button button--primary">→</button>
            </form>
          </div>
        </div>
        <div class="site-footer__bottom">
          <p>{{.copyright}}</p>
          <nav class="site-footer__legal">
            {{range .legal_links}}
            <a href="{{.url}}">{{.name}}</a>
            {{end}}
          </nav>
        </div>
      </div>
    </footer>',
    '{"brand_name": "string", "tagline": "string", "social_links": "array", "categories": "array", "company_links": "array", "newsletter_title": "string", "newsletter_description": "string", "email_placeholder": "string", "copyright": "string", "legal_links": "array"}',
    100
)
    ON CONFLICT (name) DO UPDATE SET
    html_template = EXCLUDED.html_template,
                              input_schema = EXCLUDED.input_schema,
                              updated_at = now();

-- ============================================================================
-- 6. CONTENT SITE CSS THEME
-- ============================================================================

INSERT INTO css_themes (name, display_name, description, category, semantic_tags, css_content)
VALUES (
           'content-modern',
           'Modern Content',
           'Clean, readable theme optimized for content sites and publishing',
           'content',
           ARRAY['content', 'publishing', 'readable', 'light-mode', 'minimal'],
           ':root {
             --color-primary: #2563eb;
             --color-primary-hover: #1d4ed8;
             --color-primary-text: #ffffff;
             --color-secondary: #64748b;
             --color-secondary-hover: #475569;
             --color-secondary-text: #ffffff;
             --color-accent: #f59e0b;

             --color-text: #1e293b;
             --color-text-muted: #64748b;
             --color-heading: #0f172a;
             --color-background: #ffffff;
             --color-surface: #f8fafc;
             --color-border: #e2e8f0;

             --color-header-bg: #ffffff;
             --color-header-text: #0f172a;
             --color-card-bg: #ffffff;
             --color-footer-bg: #0f172a;
             --color-footer-text: #e2e8f0;

             --border-radius: 0.5rem;
             --shadow: 0 1px 3px rgba(0,0,0,0.1);
             --shadow-lg: 0 4px 12px rgba(0,0,0,0.1);

             --font-body: "Inter", -apple-system, sans-serif;
             --font-heading: "Inter", -apple-system, sans-serif;
             --font-size-base: 1rem;
             --line-height-body: 1.7;
             --line-height-heading: 1.3;
           }

           body {
             font-family: var(--font-body);
             font-size: var(--font-size-base);
             line-height: var(--line-height-body);
           }

           /* Content site specific */
           .site-header--content {
             border-bottom: 1px solid var(--color-border);
             box-shadow: none;
           }

           .site-header__categories {
             display: flex;
             gap: 1.5rem;
             list-style: none;
           }

           .site-header__category {
             color: var(--color-text-muted);
             text-decoration: none;
             font-weight: 500;
             transition: color 0.2s;
           }

           .site-header__category:hover {
             color: var(--color-primary);
           }

           .featured-article {
             display: grid;
             grid-template-columns: 1.5fr 1fr;
             gap: 3rem;
             align-items: center;
           }

           .featured-article__image {
             position: relative;
             border-radius: var(--border-radius);
             overflow: hidden;
           }

           .featured-article__image img {
             width: 100%;
             height: 400px;
             object-fit: cover;
           }

           .featured-article__category {
             position: absolute;
             top: 1rem;
             left: 1rem;
             background: var(--color-primary);
             color: white;
             padding: 0.25rem 0.75rem;
             border-radius: 2rem;
             font-size: 0.75rem;
             font-weight: 600;
             text-transform: uppercase;
           }

           .featured-article__title {
             font-size: 2.5rem;
             line-height: var(--line-height-heading);
             margin-bottom: 1rem;
           }

           .featured-article__excerpt {
             font-size: 1.125rem;
             color: var(--color-text-muted);
             margin-bottom: 1.5rem;
           }

           .featured-article__meta {
             display: flex;
             gap: 1rem;
             color: var(--color-text-muted);
             font-size: 0.875rem;
             margin-bottom: 1.5rem;
           }

           .article-card {
             background: var(--color-card-bg);
             border-radius: var(--border-radius);
             overflow: hidden;
             box-shadow: var(--shadow);
           }

           .article-card__image {
             position: relative;
           }

           .article-card__image img {
             width: 100%;
             height: 200px;
             object-fit: cover;
           }

           .article-card__category {
             position: absolute;
             bottom: 0.75rem;
             left: 0.75rem;
             background: var(--color-primary);
             color: white;
             padding: 0.125rem 0.5rem;
             border-radius: 2rem;
             font-size: 0.625rem;
             font-weight: 600;
             text-transform: uppercase;
           }

           .article-card__content {
             padding: 1.25rem;
           }

           .article-card__title {
             font-size: 1.125rem;
             line-height: var(--line-height-heading);
             margin-bottom: 0.5rem;
           }

           .article-card__excerpt {
             color: var(--color-text-muted);
             font-size: 0.875rem;
             margin-bottom: 0.75rem;
             display: -webkit-box;
             -webkit-line-clamp: 2;
             -webkit-box-orient: vertical;
             overflow: hidden;
           }

           .article-card__meta {
             display: flex;
             gap: 1rem;
             color: var(--color-text-muted);
             font-size: 0.75rem;
           }

           .article-card--compact .article-card__image img {
             height: 140px;
           }

           .article-card--compact .article-card__content {
             padding: 1rem;
           }

           .article-card--compact .article-card__title {
             font-size: 1rem;
           }

           /* Sidebar */
           .sidebar {
             display: flex;
             flex-direction: column;
             gap: 2rem;
           }

           .sidebar__widget {
             background: var(--color-surface);
             padding: 1.5rem;
             border-radius: var(--border-radius);
           }

           .sidebar__title {
             font-size: 1rem;
             margin-bottom: 1rem;
             padding-bottom: 0.75rem;
             border-bottom: 2px solid var(--color-primary);
           }

           .popular-list {
             list-style: none;
           }

           .popular-list__item {
             padding: 0.75rem 0;
             border-bottom: 1px solid var(--color-border);
           }

           .popular-list__item:last-child {
             border-bottom: none;
           }

           .popular-list__link {
             display: flex;
             gap: 1rem;
             align-items: flex-start;
             text-decoration: none;
             color: var(--color-text);
           }

           .popular-list__number {
             font-size: 1.5rem;
             font-weight: 700;
             color: var(--color-primary);
             line-height: 1;
           }

           .popular-list__title {
             font-size: 0.875rem;
             line-height: 1.4;
           }

           .category-list {
             list-style: none;
           }

           .category-list__link {
             display: flex;
             justify-content: space-between;
             padding: 0.5rem 0;
             text-decoration: none;
             color: var(--color-text);
           }

           .category-list__count {
             color: var(--color-text-muted);
           }

           /* Ad zones */
           .ad-zone {
             background: var(--color-surface);
             border: 1px dashed var(--color-border);
             border-radius: var(--border-radius);
             padding: 1rem;
             text-align: center;
             min-height: 250px;
             display: flex;
             align-items: center;
             justify-content: center;
           }

           .ad-zone__label {
             font-size: 0.625rem;
             text-transform: uppercase;
             color: var(--color-text-muted);
             letter-spacing: 0.1em;
           }

           .ad-zone--inline {
             margin: 2rem 0;
             min-height: 100px;
           }

           /* Newsletter form */
           .newsletter-form {
             display: flex;
             flex-direction: column;
             gap: 0.75rem;
           }

           .newsletter-form__input {
             padding: 0.75rem 1rem;
             border: 1px solid var(--color-border);
             border-radius: var(--border-radius);
             font-size: 0.875rem;
           }

           .newsletter-form--footer {
             flex-direction: row;
           }

           .newsletter-form--footer .newsletter-form__input {
             flex: 1;
           }

           /* Section header with link */
           .section__header--with-link {
             display: flex;
             justify-content: space-between;
             align-items: center;
             margin-bottom: 2rem;
           }

           .section__link {
             color: var(--color-primary);
             text-decoration: none;
             font-weight: 500;
           }

           @media (max-width: 768px) {
             .featured-article {
               grid-template-columns: 1fr;
             }

             .featured-article__title {
               font-size: 1.75rem;
             }

             .site-header__categories {
               display: none;
             }
           }'
       )
    ON CONFLICT (name) DO UPDATE SET
    css_content = EXCLUDED.css_content,
                              semantic_tags = EXCLUDED.semantic_tags,
                              updated_at = now();

-- ============================================================================
-- 7. UPDATED WORKFLOW WITH CONDITIONAL ARCHITECT ROUTING
-- ============================================================================
-- Uses conditional_call_agent to route to the appropriate architect in one step

/*UPDATE agent_group_definitions
SET
    updated_at = now(),
    agent_configs = '[
    {"role": "briefer", "agent_type": "briefing-agent"},
    {"role": "chief_strategist", "agent_type": "chief-strategist"},
    {"role": "architect", "agent_type": "landing-page-architect"},
    {"role": "content_creator", "agent_type": "content-creator"},
    {"role": "html_assembler", "agent_type": "html-assembler"},
    {"role": "deployer", "agent_type": "deployer-agent"}
  ]'::jsonb,
  orchestration_workflow = '{
    "start_step": "spawn_briefer",
    "steps": {
      "spawn_briefer": {
        "action": "spawn_agent",
        "config": {"role": "briefer", "agent_type": "briefing-agent"},
        "next_step": "spawn_strategist",
        "description": "Spawn Briefing Agent"
      },
      "spawn_strategist": {
        "action": "spawn_agent",
        "config": {"role": "chief_strategist", "agent_type": "chief-strategist"},
        "next_step": "spawn_content_creator",
        "description": "Spawn Chief Strategist"
      },
      "spawn_content_creator": {
        "action": "spawn_agent",
        "config": {"role": "content_creator", "agent_type": "content-creator"},
        "next_step": "spawn_assembler",
        "description": "Spawn Content Creator"
      },
      "spawn_assembler": {
        "action": "spawn_agent",
        "config": {"role": "html_assembler", "agent_type": "html-assembler"},
        "next_step": "spawn_deployer",
        "description": "Spawn HTML Assembler"
      },
      "spawn_deployer": {
        "action": "spawn_agent",
        "config": {"role": "deployer", "agent_type": "deployer-agent"},
        "next_step": "call_briefer",
        "description": "Spawn Deployer"
      },
      "call_briefer": {
        "action": "call_agent",
        "description": "Get structured brief with site_type detection",
        "config": {
          "agent_type": "briefing-agent",
          "target_role": "briefer",
          "timeout_seconds": 60
        },
        "output_field": "brief_data",
        "next_step": "call_strategist"
      },
      "call_strategist": {
        "action": "call_agent",
        "description": "Get the Build Plan from the Strategist",
        "config": {
          "agent_type": "chief-strategist",
          "target_role": "chief_strategist",
          "input_fields": ["brief_data", "input_data"],
          "timeout_seconds": 120
        },
        "output_field": "build_plan_data",
        "next_step": "call_architect"
      },
      "call_architect": {
        "action": "conditional_call_agent",
        "description": "Route to and call appropriate architect based on site_type",
        "config": {
          "field_path": "brief_data.structured_brief.result.site_type",
          "agent_mapping": {
            "landing": "landing-page-architect",
            "content": "content-site-architect",
            "portfolio": "portfolio-architect"
          },
          "default_agent": "landing-page-architect",
          "input_fields": ["build_plan_data", "brief_data", "input_data"],
          "timeout_seconds": 120
        },
        "output_field": "template_data",
        "next_step": "call_content_creator"
      },
      "call_content_creator": {
        "action": "call_agent",
        "description": "Generate content JSON",
        "config": {
          "agent_type": "content-creator",
          "target_role": "content_creator",
          "input_fields": ["template_data", "build_plan_data", "brief_data", "input_data"],
          "timeout_seconds": 300
        },
        "output_field": "content_data",
        "next_step": "call_assembler"
      },
      "call_assembler": {
        "action": "call_agent",
        "description": "Assemble final HTML with CSS and JS",
        "config": {
          "agent_type": "html-assembler",
          "target_role": "html_assembler",
          "input_fields": ["content_data", "template_data", "brief_data", "input_data"],
          "timeout_seconds": 120
        },
        "output_field": "final_html_data",
        "next_step": "call_deployer"
      },
      "call_deployer": {
        "action": "call_agent",
        "description": "Push the final site to Git",
        "config": {
          "agent_type": "deployer-agent",
          "target_role": "deployer",
          "input_fields": ["final_html_data", "input_data"],
          "timeout_seconds": 180
        },
        "next_step": "complete"
      },
      "complete": {
        "action": "complete_workflow",
        "description": "Site build is complete."
      }
    }
  }'::jsonb,
  description = '[MVP v3] Multi-architect workflow: Routes to specialist architects (landing/content/portfolio) based on site_type'
WHERE group_type = 'mvp-site-builder';*/

===
---

updated
-------

-- ============================================================================
-- UPDATE ARCHITECT AGENTS - Fix build_plan_field configuration
-- ============================================================================

-- Fix landing-page-architect to look for "build_plan" instead of "build_plan_data"
UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = '{
    "workflow": {
      "start_step": "assemble_template",
      "steps": {
        "assemble_template": {
          "action": "assemble_from_library",
          "config": {
            "build_plan_field": "build_plan"
          },
          "next_step": "complete",
          "description": "Build the site template using component library"
        },
        "complete": {
          "action": "complete_workflow",
          "description": "Return the empty template and content requirements"
        }
      }
    },
    "processing_mode": "task",
    "timeout_seconds": 180
  }'::jsonb
WHERE type = 'landing-page-architect';

-- Fix content-site-architect to look for "build_plan" instead of "build_plan_data"
UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = '{
    "workflow": {
      "start_step": "assemble_template",
      "steps": {
        "assemble_template": {
          "action": "assemble_from_library",
          "config": {
            "build_plan_field": "build_plan"
          },
          "next_step": "complete",
          "description": "Build the content site template using component library"
        },
        "complete": {
          "action": "complete_workflow",
          "description": "Return the template and content requirements"
        }
      }
    },
    "processing_mode": "task",
    "timeout_seconds": 180
  }'::jsonb
WHERE type = 'content-site-architect';


====

updated paths
-- ============================================================================
-- UPDATE ARCHITECT AGENTS - Fix build_plan_field configuration
-- ============================================================================

-- Fix landing-page-architect to look for "build_plan" instead of "build_plan_data"
UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = '{
    "workflow": {
      "start_step": "assemble_template",
      "steps": {
        "assemble_template": {
          "action": "assemble_from_library",
          "config": {
            "build_plan_field": "build_plan"
          },
          "next_step": "complete",
          "description": "Build the site template using component library"
        },
        "complete": {
          "action": "complete_workflow",
          "description": "Return the empty template and content requirements"
        }
      }
    },
    "processing_mode": "task",
    "timeout_seconds": 180
  }'::jsonb
WHERE type = 'landing-page-architect';

-- Fix content-site-architect to look for "build_plan" instead of "build_plan_data"
UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = '{
    "workflow": {
      "start_step": "assemble_template",
      "steps": {
        "assemble_template": {
          "action": "assemble_from_library",
          "config": {
            "build_plan_field": "build_plan"
          },
          "next_step": "complete",
          "description": "Build the content site template using component library"
        },
        "complete": {
          "action": "complete_workflow",
          "description": "Return the template and content requirements"
        }
      }
    },
    "processing_mode": "task",
    "timeout_seconds": 180
  }'::jsonb
WHERE type = 'content-site-architect';

-- ============================================================================
-- UPDATE SITE-DEPLOYER - Fix field paths based on actual data structure
-- Structure is:
--   input_data.domain
--   input_data.final_html.assemble_html.final_html
-- ============================================================================

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = '{
    "workflow": {
      "start_step": "commit_to_git",
      "steps": {
        "commit_to_git": {
          "action": "git_commit",
          "config": {
            "repo_name": "sites",
            "domain_field": "input_data.domain",
            "content_field": "final_html.assemble_html.final_html",
            "commit_message": "Update site: {{.input_data.domain}}",
            "filename": "index.html"
          },
          "next_step": "complete",
          "description": "Commit to sites repo"
        },
        "complete": {
          "action": "complete_workflow",
          "description": "Deployment complete"
        }
      }
    },
    "processing_mode": "task",
    "timeout_seconds": 180
  }'::jsonb
WHERE type = 'site-deployer';

==
max tokens up, sentence length down

UPDATE agent_definitions
SET
    updated_at = now(),
    default_config = '{
    "workflow": {
      "start_step": "generate_build_plan",
      "steps": {
        "generate_build_plan": {
          "action": "execute_llm_prompt",
          "config": {
            "ai_service": {
              "provider": "anthropic",
              "model": "claude-haiku-4-5-20251001",
              "api_key_env_var": "ANTHROPIC_API_KEY",
              "max_tokens": 8000
            },
            "input_fields": ["input_data", "brief_data"],
            "output_field": "build_plan_json",
            "prompt_template": "You are a website strategist creating a Build Plan based on behavioral psychology.\n\nWebsite Request:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Model: {{.input_data.model}}\n\n{{if .brief_data}}Brief Data:\n{{.brief_data}}\n{{end}}\n\nBehavioral Models:\n- AIDA: Attention → Interest → Desire → Action\n- PAS: Problem → Agitate → Solution\n- FAB: Features → Advantages → Benefits\n- 4Ps: Promise → Picture → Proof → Push\n\nAvailable Components: header, hero, features, social_proof, pricing, faq, call_to_action, footer\n\nCreate a build plan that maps the behavioral model to sections with specific guidance.\n\nReturn ONLY valid JSON:\n{\n  \"model\": \"AIDA\",\n  \"sections\": [\"header\", \"hero\", \"features\", \"social_proof\", \"pricing\", \"faq\", \"call_to_action\", \"footer\"],\n  \"section_guidance\": {\n    \"hero\": {\n      \"stage\": \"Attention\",\n      \"purpose\": \"Grab attention with bold value proposition\",\n      \"key_message\": \"Main benefit headline\",\n      \"tone\": \"Confident, clear\"\n    },\n    \"features\": {\n      \"stage\": \"Interest\",\n      \"purpose\": \"Build interest with capabilities\",\n      \"key_message\": \"What it does\",\n      \"tone\": \"Informative\"\n    }\n  },\n  \"theme\": \"tech-saas\"\n}\n\nProvide section_guidance for each section in the sections array. Keep guidance concise (1-2 sentences each). Return ONLY the JSON object."
          },
          "next_step": "complete"
        },
        "complete": {
          "action": "complete_workflow",
          "description": "Return the Build Plan"
        }
      }
    },
    "processing_mode": "task",
    "timeout_seconds": 120
  }'::jsonb
WHERE type = 'site-strategist';