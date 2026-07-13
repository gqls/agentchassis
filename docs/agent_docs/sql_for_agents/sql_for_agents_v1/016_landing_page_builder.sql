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