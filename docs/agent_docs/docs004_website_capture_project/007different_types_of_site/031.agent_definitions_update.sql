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
                     "prompt_template": "You are a professional website content creator. Your job is to create compelling, industry-specific CONTENT (not HTML structure).\n\nWebsite Details:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n\nBrief Data (from questionnaire):\n{{.brief_data}}\n\nBuild Strategy (from strategist):\n{{.build_plan}}\n\nContent Requirements - these are the placeholders you need to fill:\n{{.template_data.content_requirements}}\n\nYour Task:\nCreate a JSON object with content for each placeholder. Group by component.\n\nGuidelines:\n- Write compelling, conversion-focused copy\n- Use the brief data to inform tone, messaging, and specifics\n- Match the domain and industry tone\n- For testimonials, use optimistic placeholder attributions like \"[Future You]\", \"[Soon-to-be Satisfied Customer]\" - NOT fake names\n- Use action-oriented language for CTAs\n- Keep brand consistency throughout\n- Stats/numbers should be realistic placeholders like \"500+\" or \"10,000+\"\n- If brief contains brand_name, use it consistently\n- If brief contains primary_cta, use that exact text for main call-to-action\n\nReturn ONLY valid JSON in this exact structure:\n{\n  \"meta\": {\n    \"title\": \"Page title for browser tab\",\n    \"description\": \"SEO meta description (150-160 chars)\"\n  },\n  \"theme\": \"recommended theme name from: default, clean-minimal, bold-conversion, warm-friendly, tech-saas, luxury-premium\",\n  \"theme_tags\": [\"semantic\", \"tags\", \"for\", \"theme\", \"matching\"],\n  \"sections\": {\n    \"component_header_0\": {\n      \"brand_name\": \"Your Brand Name\",\n      \"cta_text\": \"CTA Button Text\"\n    },\n    \"component_hero_1\": {\n      \"headline\": \"Main headline\",\n      \"subheadline\": \"Supporting text\",\n      \"primary_cta\": \"Primary button\",\n      \"secondary_cta\": \"Secondary button\"\n    }\n  }\n}\n\nFill ALL placeholders from the content requirements. Return ONLY the JSON object, no markdown or explanation."
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