SELECT * FROM agent_definitions;
-[ RECORD 1 ]-+--
id                     | e221ac1e-69d5-46e9-b2c6-80e426abecb2
type                   | site-component-architect
display_name           | Site Architect Agent for Component Assembly
description            | Assembles empty HTML templates from the in-house component library.
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow"}, "assemble_template": {"action": "assemble_from_library", "config": {"input_fields": ["build_plan_data"]}, "next_step": "complete"}}, "start_step": "assemble_template"}, "processing_mode": "task", "timeout_seconds": 120}
is_active              | t
created_at             | 2025-11-15 17:44:27.186739+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["build", "assemble", "database"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | executor
status                 | active
domain_tags            | ["website"]
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 2 ]-+--
id                     | 60195d6d-3960-41cb-9e2a-ae51a5aadeb8
type                   | site-classifier
display_name           | Site Classifier
description            | Analyzes domain and objective to determine site type and recommend appropriate builder group
category               | classification
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return classification result"}, "classify_site": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "max_tokens": 1500, "api_key_env_var": "ANTHROPIC_API_KEY"}, "output_type": "json", "input_fields": ["input_data", "available_builders"], "output_field": "classification_result", "prompt_template": "Classify this website project and recommend the appropriate builder.\n\nInput:\n- Domain: {{.domain}}\n- Objective: {{.objective}}\n\nAvailable Builders:\n{{range .available_builders.agents}}- {{.type}}: {{.description}}\n{{end}}\n\nClassify the site into ONE of these types based on the objective:\n\n**landing** - Conversion-focused single-purpose sites:\n- Product/service sales pages, SaaS landing pages\n- Lead generation, signups, app downloads\n- Event registration, clear single CTA goal\n\n**content** - Publishing/content sites:\n- News, blogs, magazines, articles\n- Content aggregation, SEO/traffic focused\n- Category navigation, archives\n\n**portfolio** - Showcase/portfolio sites:\n- Creative portfolios, agencies, case studies\n- Visual/image heavy, project galleries\n\n**brochure** - Multi-page business sites:\n- Corporate sites, general business presence\n- Service providers, consultants, professional services\n- About/Services/Contact structure\n\nReturn ONLY valid JSON with this structure:\n{\n  \"site_type\": \"landing|content|portfolio|brochure\",\n  \"confidence\": 0.0-1.0,\n  \"reasoning\": \"brief explanation\",\n  \"recommended_builder\": \"builder-type\",\n  \"detected_industry\": \"industry name\",\n  \"detected_signals\": [\"signal1\", \"signal2\", ...]\n}"}, "next_step": "complete"}}, "start_step": "classify_site"}, "processing_mode": "task", "timeout_seconds": 30}
is_active              | t
created_at             | 2025-11-28 10:01:15.964026+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["classification", "analysis", "llm"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "250m", "memory": "256Mi"}, "requests": {"cpu": "50m", "memory": "128Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 3 ]-+--
id                     | 880e80e8-f303-43c3-bf69-8462b94678f4
type                   | chief-strategist
display_name           | Chief Strategist Agent
description            | Creates a "first-principles" Build Plan (e.g., AIDA, PAS) from a simple objective.
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return parsed plan"}, "parse_plan": {"action": "parse_json_field", "config": {"source_field": "build_plan_raw"}, "next_step": "complete", "description": "Parse JSON from LLM response using existing datahelpers", "output_field": "plan_data"}, "generate_build_plan": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "max_tokens": 8192, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_data": ["domain", "objective", "model"], "prompt_template": "You are a Site Planner designing the structure for {{.input_data.domain}}.\n\nOBJECTIVE: {{.input_data.objective}}\nMARKETING MODEL: {{.input_data.model}}\n\nSTEP 1: Determine the best site type for this objective.\n\nSite Type Guidelines:\n\nLANDING (1 page, 5-8 components)\n- Product launches, lead generation, focused campaigns\n- Single conversion goal, minimal navigation\n- Revenue: Direct sales, lead capture\n\nCORPORATE (4-6 pages)\n- Professional services, consulting, established businesses\n- Trust-building, multiple service areas\n- Revenue: Service contracts, B2B relationships\n\nPORTFOLIO (3-5 pages)\n- Creatives, agencies, freelancers\n- Case study focused, visual showcase\n- Revenue: Project work, client acquisition\n\nECOMMERCE (2-4 pages + product structure)\n- Product sales, shopping focused\n- Category browsing, cart functionality\n- Revenue: Direct product sales\n\nCONTENT (4-8 pages + article structure)\n- News sites, blogs, recipes, celebrity gossip, lifestyle\n- Content-driven traffic, regular publishing\n- SEO focused, high page count potential\n- Revenue: Advertising, affiliate links, sponsored content\n\nTOOLS (2-5 pages + tool interfaces)\n- Calculators (mortgage, tiles, BMI, etc.), converters, utilities\n- Feature/functionality driven, practical value\n- User retention through bookmarking\n- Revenue: Advertising, affiliate referrals, premium features\n\nSTEP 2: Plan each page with specific components.\n\nAvailable component types:\n- hero-centered, hero-split, hero-video\n- services-grid, services-list\n- features-cards, features-comparison\n- testimonials-carousel, testimonials-grid\n- team-grid, pricing-tiers, faq-accordion\n- cta-banner, cta-split\n- contact-form, contact-simple\n- about-story, about-values\n- footer-standard\n- blog-grid, blog-featured, article-layout\n- recipe-card, recipe-grid, recipe-detail\n- tool-calculator, tool-converter, tool-interface\n- ad-banner, ad-sidebar, affiliate-showcase\n- category-grid, content-feed, search-bar\n- social-share, comments-section, newsletter-signup\n\nSTEP 3: Create the sitemap with navigation structure.\n\nThe sitemap defines:\n- How each page appears in navigation\n- The URL for each page\n- Whether it appears in header nav, footer nav, or both\n\nURL Rules:\n- Home page: /index.html\n- Other pages: /{page-name}.html (e.g., /about.html, /services.html)\n- Use lowercase, hyphens for multi-word names\n\nOUTPUT FORMAT (valid JSON only):\n{\n  \"site_type\": \"landing|corporate|portfolio|ecommerce|content|tools\",\n  \"reasoning\": \"Why this structure fits the objective\",\n  \"theme_suggestion\": \"professional|bold|minimal|creative|editorial|functional\",\n  \"revenue_model\": \"direct_sales|services|advertising|affiliate|freemium\",\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Brand\",\n      \"purpose\": \"What this page achieves\",\n      \"components\": [\n        {\"type\": \"hero-centered\", \"priority\": \"high\"},\n        {\"type\": \"services-grid\", \"priority\": \"high\"}\n      ],\n      \"meta_description\": \"SEO description\"\n    }\n  ],\n  \"sitemap\": [\n    {\"label\": \"Home\", \"page\": \"index\", \"url\": \"/index.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"About\", \"page\": \"about\", \"url\": \"/about.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Services\", \"page\": \"services\", \"url\": \"/services.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Contact\", \"page\": \"contact\", \"url\": \"/contact.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Privacy Policy\", \"page\": \"privacy\", \"url\": \"/privacy.html\", \"in_header\": false, \"in_footer\": true}\n  ],\n  \"global\": {\n    \"brand_tone\": \"professional|friendly|bold|technical|editorial|practical\",\n    \"primary_cta\": {\"text\": \"Get Started\", \"url\": \"/contact.html\"}\n  }\n}\n\nIMPORTANT:\n- Every page in \"pages\" array MUST have a corresponding entry in \"sitemap\"\n- The \"page\" field in sitemap MUST match the \"name\" field in pages\n- Home page is always /index.html\n- Include privacy/terms pages in footer only (in_header: false)\n\nEXAMPLE - Corporate site:\n{\n  \"site_type\": \"corporate\",\n  \"pages\": [\n    {\"name\": \"index\", \"title\": \"Home | Acme Corp\", \"components\": [{\"type\": \"hero-centered\"}, {\"type\": \"services-grid\"}, {\"type\": \"testimonials-carousel\"}, {\"type\": \"cta-banner\"}]},\n    {\"name\": \"about\", \"title\": \"About Us | Acme Corp\", \"components\": [{\"type\": \"about-story\"}, {\"type\": \"team-grid\"}, {\"type\": \"about-values\"}]},\n    {\"name\": \"services\", \"title\": \"Our Services | Acme Corp\", \"components\": [{\"type\": \"hero-split\"}, {\"type\": \"services-list\"}, {\"type\": \"faq-accordion\"}]},\n    {\"name\": \"contact\", \"title\": \"Contact Us | Acme Corp\", \"components\": [{\"type\": \"contact-form\"}, {\"type\": \"cta-split\"}]}\n  ],\n  \"sitemap\": [\n    {\"label\": \"Home\", \"page\": \"index\", \"url\": \"/index.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"About Us\", \"page\": \"about\", \"url\": \"/about.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Services\", \"page\": \"services\", \"url\": \"/services.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Contact\", \"page\": \"contact\", \"url\": \"/contact.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Privacy Policy\", \"page\": \"privacy\", \"url\": \"/privacy.html\", \"in_header\": false, \"in_footer\": true}\n  ],\n  \"global\": {\n    \"brand_tone\": \"professional\",\n    \"primary_cta\": {\"text\": \"Get in Touch\", \"url\": \"/contact.html\"}\n  }\n}"}, "next_step": "parse_plan", "description": "Create the Build Plan using LLM", "output_field": "build_plan_raw"}}, "start_step": "generate_build_plan"}, "processing_mode": "task", "timeout_seconds": 120}
is_active              | t
created_at             | 2025-11-15 18:27:24.625309+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["strategy", "llm", "planning"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | strategist
status                 | active
domain_tags            | ["website"]
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 4 ]-+--
id                     | fec422b0-25a2-401f-83e3-e3ce9e76682f
type                   | reasoning
display_name           | Reasoning Agent
description            | Performs logical analysis and decision making
category               | code-driven
default_config         | {"model": "claude-3-opus", "ai_service": {"model": "claude-3-5-sonnet", "provider": "anthropic"}, "temperature": 0.2, "processing_mode": "task"}
is_active              | t
created_at             | 2025-08-10 16:00:36.153034+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | []
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"dlq": "system.agent.{type}.dlq", "errors": "system.agent.{type}.errors", "requests": "system.agent.{type}.requests", "responses": "system.agent.{type}.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | {"steps": {"execute": {"action": "execute_llm_prompt", "next_step": "complete", "description": "Execute task"}, "complete": {"action": "complete_workflow", "description": "Complete"}}, "start_step": "execute"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 5 ]-+--
id                     | 6f40f1f5-34d1-4892-b78b-f08e042f4898
type                   | briefing-agent
display_name           | Briefing Agent
description            | Executes briefing questionnaires - either via LLM inference or HITL collection
category               | data-collection
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return the completed brief"}, "check_mode": {"action": "evaluate_condition", "config": {"default": "infer_via_llm", "conditions": {"auto": "infer_via_llm", "interactive": "collect_via_hitl"}, "condition_field": "input_data.hitl_mode"}, "next_step": "infer_via_llm", "description": "Determine if briefing should be interactive or auto-inferred"}, "infer_via_llm": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "max_tokens": 4000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["input_data", "questionnaire"], "output_field": "brief_answers", "prompt_template": "You are completing a briefing questionnaire for a website project.\n\nProject Info:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Site Type: {{.input_data.site_type}}\n- Industry: {{.input_data.detected_industry}}\n\nQuestionnaire to complete:\n{{.questionnaire}}\n\nBased on the domain name, objective, and your knowledge of the industry, provide thoughtful answers to each question in the questionnaire.\n\nFor questions you cannot confidently answer from the available information, provide reasonable defaults appropriate for the industry.\n\nReturn your answers as a JSON object where keys match the field names in the questionnaire:\n{\n  \"field_name_1\": \"your answer\",\n  \"field_name_2\": \"your answer\",\n  ...\n}\n\nReturn ONLY valid JSON."}, "next_step": "complete"}, "collect_via_hitl": {"action": "request_human_input", "config": {"message": "Please complete the briefing questionnaire for this project", "request_type": "questionnaire", "timeout_seconds": 86400, "questionnaire_field": "questionnaire"}, "next_step": "complete", "description": "Collect answers via human-in-the-loop", "output_field": "brief_answers"}}, "start_step": "check_mode"}, "processing_mode": "task", "timeout_seconds": 120}
is_active              | t
created_at             | 2025-11-28 09:55:45.098946+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["briefing", "questionnaire", "llm", "hitl"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 6 ]-+--
id                     | c3968560-f16f-4441-86f2-aee956f49b5c
type                   | content-writer
display_name           | Content Writer
description            | Creates website content from brief data and template requirements. Works with intake orchestrator flow.
category               | content
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return the content JSON"}, "generate_content": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "max_tokens": 8000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["template_data", "build_plan", "brief_data", "input_data"], "output_field": "content_json", "prompt_template": "You are a professional website content creator. Your job is to create compelling, industry-specific CONTENT.\n\nWebsite Details:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n\nBuild Strategy (for tone and messaging guidance):\n{{.build_plan}}\n\nContent Requirements - THESE ARE THE EXACT PLACEHOLDERS YOU MUST FILL:\n{{.template_data.assemble_template.content_requirements}}\n\nIMPORTANT: You must use the EXACT key names from content_requirements. Do not rename, abbreviate, or restructure them.\n\nFor example, if content_requirements has:\n  \"feature_1_title\": \"Fast & Reliable\"\nYour output must use \"feature_1_title\" as the key, NOT \"feature_1_headline\".\n\nGuidelines:\n- Write compelling, conversion-focused copy\n- Match the domain and industry tone\n- For testimonials, use placeholder attributions like \"[Future Customer]\" - NOT fake names\n- Stats/numbers must be truthful - do not invent metrics\n\nReturn ONLY valid JSON in this structure:\n{\n  \"meta\": {\n    \"title\": \"Page title for browser tab\",\n    \"description\": \"SEO meta description (150-160 chars)\"\n  },\n  \"theme\": \"tech-saas\",\n  \"sections\": {\n    \"component_header_0\": {\n      \"brand_name\": \"Your value here\",\n      \"cta_text\": \"Your value here\"\n    }\n  }\n}\n\nFill ALL placeholders using their EXACT key names. Return ONLY JSON, no markdown."}, "next_step": "complete"}}, "start_step": "generate_content"}, "processing_mode": "task", "timeout_seconds": 300}
is_active              | t
created_at             | 2025-11-29 18:06:38.077414+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["content", "copywriting", "llm"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | executor
status                 | active
domain_tags            | ["content", "website"]
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 7 ]-+--
id                     | d8c6adcc-7ce7-4b18-af1b-2f34db616c06
type                   | multipage-website-builder
display_name           | Multi-Page Website Builder
description            | Builds large websites (20+ pages) using batched generation to avoid token limits
category               | orchestrator
default_config         | {"workflow": {"steps": {"deploy": {"action": "call_agent", "config": {"agent_type": "deployer-agent", "target_role": "deployer", "input_fields": ["site_files", "input_data", "site_record"], "timeout_seconds": 180}, "next_step": "update_timestamps", "description": "Deploy site to git repository", "output_field": "deployment_result"}, "complete": {"action": "complete_workflow", "description": "Multipage site build complete"}, "assemble_site": {"action": "assemble_multipage_site", "config": {"pages_field": "all_pages", "add_navigation": true, "include_robots_txt": true, "include_sitemap_xml": true, "generate_standard_pages": true}, "next_step": "deploy", "description": "Assemble pages into complete site with navigation", "output_field": "site_files"}, "spawn_deployer": {"action": "spawn_agent", "config": {"role": "deployer", "agent_type": "deployer-agent"}, "next_step": "call_strategist", "description": "Spawn deployer agent", "output_field": "deployer_info"}, "call_strategist": {"action": "call_agent", "config": {"agent_type": "chief-strategist", "target_role": "strategist", "input_fields": ["input_data", "site_record"], "timeout_seconds": 120}, "next_step": "sync_pages_to_db", "description": "Get page plan from chief strategist", "output_field": "page_plan"}, "spawn_strategist": {"action": "spawn_agent", "config": {"role": "strategist", "agent_type": "chief-strategist"}, "next_step": "spawn_content_creator", "description": "Spawn chief strategist agent", "output_field": "strategist_info"}, "sync_pages_to_db": {"action": "sync_pages_to_db", "config": {"input_fields": ["site_record", "page_plan"]}, "next_step": "generate_pages_loop", "description": "Sync pages to database and build navigation structure", "output_field": "db_sync"}, "update_timestamps": {"action": "update_site_timestamps", "config": {"input_fields": ["site_record"]}, "next_step": "complete", "description": "Update site last_built_at and last_deployed_at", "output_field": "timestamp_update"}, "ensure_site_record": {"action": "ensure_site_record", "config": {"input_fields": ["input_data"]}, "next_step": "spawn_strategist", "description": "Create or retrieve site record from database", "output_field": "site_record"}, "generate_pages_loop": {"action": "loop", "config": {"loop_var": "current_page", "substeps": {"create_html": {"action": "call_agent", "config": {"agent_type": "html-developer", "target_role": "developer", "input_fields": ["page_content", "current_page", "input_data", "db_sync", "page_plan"], "timeout_seconds": 180}, "next_step": "extract_links", "description": "Convert content to professional HTML page", "output_field": "page_html"}, "extract_links": {"action": "extract_and_sync_links", "config": {"input_fields": ["page_html", "current_page", "site_record"]}, "description": "Extract links from HTML and sync to link registry", "output_field": "link_sync"}, "generate_content": {"action": "call_agent", "config": {"agent_type": "content-creator", "target_role": "writer", "input_fields": ["current_page", "input_data", "page_plan"], "timeout_seconds": 180}, "next_step": "create_html", "description": "Generate content strategy/copy for page", "output_field": "page_content"}}, "iterate_over": "page_plan.response.plan_data.pages", "max_iterations": 20}, "next_step": "assemble_site", "description": "Generate all pages with content and HTML conversion", "output_field": "all_pages"}, "spawn_html_developer": {"action": "spawn_agent", "config": {"role": "developer", "agent_type": "html-developer"}, "next_step": "spawn_deployer", "description": "Spawn HTML developer to convert content to pages", "output_field": "developer_info"}, "spawn_content_creator": {"action": "spawn_agent", "config": {"role": "writer", "agent_type": "content-creator"}, "next_step": "spawn_html_developer", "description": "Spawn content creator agent for loop iterations", "output_field": "writer_info"}}, "start_step": "ensure_site_record"}, "timeout_seconds": 600}
is_active              | t
created_at             | 2025-12-06 19:02:26.358753+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["orchestration", "website-builder", "multi-page", "batched-generation"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {"sections": [{"name": "company_info", "title": "Company Information", "questions": [{"type": "text", "field": "company_name", "label": "Company Name", "required": true}, {"type": "textarea", "field": "about_us", "label": "About Us", "required": true}, {"type": "text", "field": "tagline", "label": "Tagline", "required": false}]}, {"name": "services", "title": "Services & Offerings", "questions": [{"type": "json_array", "field": "services", "label": "Services (list of {name, description})", "required": true}]}, {"name": "team", "title": "Team & Leadership", "questions": [{"type": "json_array", "field": "leadership_team", "label": "Leadership Team (list of {name, title, bio})", "required": false}]}, {"name": "portfolio", "title": "Case Studies & Portfolio", "questions": [{"type": "json_array", "field": "case_studies", "label": "Case Studies", "required": false}]}, {"name": "contact", "title": "Contact Information", "questions": [{"type": "text", "field": "contact_email", "label": "Email", "required": true}, {"type": "text", "field": "contact_phone", "label": "Phone", "required": false}, {"type": "text", "field": "headquarters", "label": "Location", "required": false}]}, {"name": "features", "title": "Site Features", "questions": [{"type": "boolean", "field": "has_blog", "label": "Include Blog?", "default": false}, {"type": "boolean", "field": "has_careers", "label": "Include Careers?", "default": false}]}]}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 8 ]-+--
id                     | 036c2428-a0fd-4c75-ad9c-4b1f9fe843e2
type                   | site-component-architect
display_name           | Site Architect Agent
description            | Assembles empty HTML templates.
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow"}, "assemble_template": {"action": "assemble_from_library", "config": {"input_fields": ["build_plan_data"]}, "next_step": "complete"}}, "start_step": "assemble_template"}, "processing_mode": "task", "timeout_seconds": 120}
is_active              | t
created_at             | 2025-11-19 13:49:55.211764+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["build"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 2
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | executor
status                 | active
domain_tags            | ["website"]
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 9 ]-+--
id                     | 7c1939a8-b237-4e1a-8edd-7d30264b9b14
type                   | deployer-agent
display_name           | Site Deployer Agent (Git)
description            | Commits final site files to a Git repository by calling the git-adapter.
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Deployment complete"}, "commit_to_git": {"action": "git_commit", "config": {"filename": "index.html", "repo_name": "sites", "domain_field": "domain", "input_fields": ["input_data"], "content_field": "input_data.final_site_data.generate_content.result", "commit_message": "Update site: {{.domain}}"}, "next_step": "complete", "description": "Commit to sites repo"}}, "start_step": "commit_to_git"}, "processing_mode": "task", "timeout_seconds": 180}
is_active              | t
created_at             | 2025-11-15 18:29:07.392599+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["git", "deploy", "adapter"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | integrator
status                 | active
domain_tags            | ["deployment", "git"]
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 10 ]---+--
id                     | 535f8d1b-5d9b-42b3-a4b7-9a1432421fef
type                   | content-creator-about
display_name           | About Page Writer
description            | Specialized in writing compelling about page content that tells the story of a business or explains a concept
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return about page content"}, "generate_about_content": {"action": "execute_llm_prompt", "config": {"input_fields": ["input_data"], "prompt_template": "{{.prompt}}\n\nProvide well-structured content suitable for an about page. Use clear paragraphs and maintain a professional yet approachable tone. As a small addendum add that the site is created by ai agents and may be for sale."}, "next_step": "complete", "description": "Generate about page content"}}, "start_step": "generate_about_content"}, "ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "max_tokens": 2000, "temperature": 0.7, "processing_mode": "task"}
is_active              | t
created_at             | 2025-10-30 20:10:19.161522+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["content", "about", "storytelling", "brand-narrative"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.content-creator-about", "process": "system.agent.content-creator-about.process", "response": "system.responses.content-creator-about"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 11 ]---+--
id                     | 80690f44-a350-401f-97ee-96ce2f840e94
type                   | content-researcher
display_name           | Content Researcher
description            | Researches and gathers comprehensive information for website content
category               | code-driven
default_config         | {"ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic"}, "processing_mode": "task"}
is_active              | t
created_at             | 2025-08-21 11:48:56.381213+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["research", "analysis", "fact-checking", "content-gathering", "competitor-analysis"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | {./agent-chassis,-config,configs/agent-chassis.yaml}
resources              | {"limits": {"cpu": "600m", "memory": "768Mi"}, "requests": {"cpu": "150m", "memory": "384Mi"}}
topics                 | {"dlq": "system.agent.content-researcher.dlq", "errors": "system.agent.content-researcher.errors", "requests": "system.agent.content-researcher.requests", "responses": "system.agent.content-researcher.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | {"steps": {"complete": {"action": "complete_workflow", "description": "Return research findings"}, "research": {"action": "execute_llm_prompt", "config": {"prompt_template": "Research content strategy for: {{.business_type}}. Include keywords, topics, and content recommendations."}, "next_step": "complete", "description": "Research content topics and keywords"}}, "start_step": "research"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": false}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 12 ]---+--
id                     | 2cb37e59-6e08-43c6-b356-495d9b784aca
type                   | research-agent
display_name           | Research Agent
description            | Researches topics via web search, extracts relevant quotes, synthesizes findings with full source attribution. Stores results in research_results table for citation.
category               | specialist
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output": {"id": "stored_research.id", "query": "search_query.result", "summary": "synthesis.result.summary", "confidence": "synthesis.result.confidence", "key_points": "synthesis.result.key_points", "source_count": "research_content.source_count", "content_quality": "research_content.content_quality", "recommendations": "synthesis.result.recommendations"}}, "description": "Complete research workflow and return results"}, "search_web": {"action": "web_search", "config": {"query_from": "search_query.result", "num_results": 10}, "next_step": "prepare_urls", "description": "Search the web for relevant sources", "output_field": "search_results"}, "synthesize": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5", "provider": "anthropic", "max_tokens": 1500, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["extracted", "research_content"], "output_format": "json", "prompt_template": "Synthesize research findings about: {{.extracted.topic}}\n\n{{.research_content.research_text}}\n\n## Task\nWrite a comprehensive synthesis of the findings. Cite sources by index [0], [1], [2] where applicable.\n\nReturn JSON:\n{\n  \"summary\": \"3-4 paragraph synthesis with [citations] where relevant\",\n  \"key_points\": [\"key point with [citation]\", \"another point [citation]\"],\n  \"recommendations\": [\"actionable recommendation based on findings\"],\n  \"confidence\": 0.0-1.0\n}\n\nRules:\n- Cite sources when making specific claims\n- Be factual and objective\n- Note any conflicting information between sources\n- Rate confidence based on source quality and agreement"}, "next_step": "store_research", "description": "Synthesize findings into comprehensive summary with citations", "output_field": "synthesis"}, "prepare_urls": {"action": "prepare_urls", "config": {"max_scrapes": 3, "max_snippets": 5, "prefer_domains": [".gov", ".edu", ".org", "reuters.com", "bbc.com", "forbes.com", "hbr.org", "mckinsey.com"], "exclude_domains": ["pinterest.com", "facebook.com", "twitter.com", "instagram.com", "youtube.com", "reddit.com"]}, "next_step": "scrape_pages", "description": "Extract top URLs for scraping and collect snippets", "output_field": "prepared_urls"}, "scrape_pages": {"action": "batch_webscrape", "config": {"urls_field": "prepared_urls.urls_to_scrape", "scrape_config": {"only_main_content": true, "capture_screenshot": false}}, "next_step": "format_content", "description": "Scrape top result pages for detailed content", "output_field": "scrape_results", "timeout_seconds": 60}, "extract_topic": {"action": "extract_fields", "config": {"fields": {"topic": ["current_section.topic", "current_section.research_query", "current_section.name"], "company": ["reviewed_brief.company_name"], "industry": ["reviewed_brief.industry", "reviewed_brief.category"]}}, "next_step": "build_search_query", "description": "Extract research topic from section data", "output_field": "extracted"}, "format_content": {"action": "format_research_content", "config": {"scrape_field": "scrape_results", "snippets_field": "prepared_urls.snippet_context", "max_content_per_source": 6000}, "next_step": "synthesize", "description": "Format scraped content and snippets for synthesis", "output_field": "research_content"}, "store_research": {"action": "insert_research_result", "config": {"table": "research_results", "fields": {"query": "search_query.result", "topic": "extracted.topic", "site_id": "site_record.site_id", "sources": "research_content.sources", "summary": "synthesis.result", "findings": "synthesis.result"}}, "next_step": "complete", "description": "Store research in database for future reference", "output_field": "stored_research"}, "build_search_query": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "max_tokens": 150, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["extracted"], "prompt_template": "Create a focused web search query to research:\n\nTopic: {{.extracted.topic}}\nIndustry: {{.extracted.industry}}\nCompany: {{.extracted.company}}\n\nReturn ONLY the search query string, 3-8 words, no quotes or operators.\nFocus on finding authoritative, recent information about this topic."}, "next_step": "search_web", "description": "Build effective search query from context", "output_field": "search_query"}}, "start_step": "extract_topic"}, "processing_mode": "task", "timeout_seconds": 180}
is_active              | t
created_at             | 2025-12-22 17:47:31.543928+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["web-search", "research", "citation", "synthesis"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | active
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"site_record": "object with site_id for storing results", "reviewed_brief": "object with industry, company context", "current_section": "object with topic or research_query field"}, "required": ["current_section"]}
output_contract        | {"produces": {"id": "uuid - research_results record ID", "query": "string - the search query used", "sources": "array of {url, title, domain, quotes[], accessed_at}", "summary": "string - synthesized findings with citations", "source_count": "number"}}
-[ RECORD 13 ]---+--
id                     | 069a5453-5248-4764-bfeb-79356ccbcea5
type                   | content-site-architect
display_name           | Content Site Architect
description            | Assembles content/publishing sites with article grids, category nav, and ad zones
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return the template and content requirements"}, "assemble_template": {"action": "assemble_from_library", "config": {"build_plan_field": "build_plan"}, "next_step": "complete", "description": "Build the content site template using component library"}}, "start_step": "assemble_template"}, "processing_mode": "task", "timeout_seconds": 180}
is_active              | t
created_at             | 2025-11-28 13:08:04.003316+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["build", "assemble", "database", "content-site", "publishing"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 14 ]---+--
id                     | 0a58cf02-4a9e-4900-9c7e-207622f12247
type                   | content-reviewer
display_name           | Content Reviewer
description            | Reviews page content for quality, accuracy, and brand alignment. Supports HITL mode (human review with edits) and auto-eval mode (LLM review with auto-approve or flag).
category               | specialist
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output": {"edits": "review_result.edits", "issues": "review_result.issues", "content": "review_result.content", "approved": "review_result.approved", "review_mode": "review_result.review_mode", "reviewed_at": "review_result.reviewed_at", "reviewed_by": "review_result.reviewed_by"}}}, "auto_eval_content": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5", "provider": "anthropic", "max_tokens": 1500, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["current_page", "page_content", "reviewed_brief"], "output_format": "json", "prompt_template": "Review this page content for quality and accuracy.\n\n## Page\nName: {{.current_page.name}}\nTitle: {{.current_page.title}}\n\n## Company Brief\nCompany: {{.reviewed_brief.company_name}}\nIndustry: {{.reviewed_brief.industry}}\nTone: {{.reviewed_brief.tone}}\nServices: {{.reviewed_brief.services}}\n\n## Content to Review\n{{if .page_content.sections}}{{range .page_content.sections}}\n### Section: {{.component_name}}\n{{.rendered_html}}\n{{end}}{{else if .page_content.process_sections_loop_complete}}{{range .page_content.process_sections_loop_complete.results}}\n### Section: {{.component_name}}\n{{.rendered_html}}\n{{end}}{{else if .page_content.processed_sections}}{{range .page_content.processed_sections.results}}\n### Section: {{.component_name}}\n{{.rendered_html}}\n{{end}}{{else}}\n### Full Page HTML\n{{.page_content.compile_page.page_html}}\n{{end}}\n\n## Evaluation Criteria\n1. Accuracy: Does content match the brief? No invented claims?\n2. Completeness: Are all sections filled in properly?\n3. Quality: Professional tone? No placeholder text?\n4. Brand Alignment: Matches company voice and values?\n5. Technical: Valid HTML? Proper structure?\n\n## Return JSON:\n```json\n{\n  \"approved\": true/false,\n  \"overall_score\": 0.0-1.0,\n  \"issues\": [\n    {\n      \"section\": \"hero\",\n      \"severity\": \"error|warning|info\",\n      \"issue\": \"Description of the issue\",\n      \"suggestion\": \"How to fix it\"\n    }\n  ],\n  \"strengths\": [\"Good point 1\", \"Good point 2\"],\n  \"summary\": \"Brief overall assessment\"\n}\n```\n\nApprove if:\n- No errors (warnings are OK)\n- Score >= 0.7\n- No placeholder text detected\n- Content matches brief"}, "next_step": "check_auto_approval", "description": "LLM evaluates content quality and accuracy", "output_field": "eval_result"}, "escalate_to_human": {"action": "request_human_input", "config": {"message": "Auto-review found issues with {{current_page.name}} - human review required", "editable": true, "ui_config": {"title": "Content Review - Issues Found", "description": "Auto-review flagged issues. Please review and fix.", "show_issues": true, "issues_field": "eval_result.issues"}, "request_type": "review", "timeout_seconds": 3600, "notification_topic": "system.notifications.ui"}, "next_step": "process_escalation_response", "description": "Escalate to human - auto-eval found issues", "output_field": "escalation_response"}, "check_auto_approval": {"action": "conditional", "config": {"condition": "eval_result.approved == true AND eval_result.overall_score >= 0.7", "else_step": "escalate_to_human", "then_step": "finalize_auto_result"}, "description": "Check if auto-eval passed or needs human review"}, "prepare_hitl_review": {"action": "prepare_review_data", "config": {"include_fields": ["current_page", "page_content", "reviewed_brief"], "format_for_display": true}, "next_step": "request_human_review", "description": "Prepare content for human review interface", "output_field": "review_data"}, "finalize_auto_result": {"action": "build_review_result", "config": {"approved": true, "reviewer": "eval-agent", "eval_score": "eval_result.overall_score", "review_mode": "auto-eval"}, "next_step": "update_component_status", "description": "Build result from successful auto-eval", "output_field": "review_result"}, "finalize_hitl_result": {"action": "build_review_result", "config": {"edits_field": "processed_response.edits", "review_mode": "hitl", "approved_field": "processed_response.approved", "reviewer_field": "processed_response.responded_by"}, "next_step": "update_component_status", "description": "Build final result from HITL review", "output_field": "review_result"}, "request_human_review": {"action": "request_human_input", "config": {"message": "Review page content for {{current_page.name}}", "editable": true, "ui_config": {"title": "Content Review", "show_diff": true, "description": "Review and edit page content before publishing", "allow_comments": true}, "data_field": "review_data", "request_type": "review", "stop_on_cancel": false, "timeout_seconds": 3600, "notification_topic": "system.notifications.ui"}, "next_step": "process_human_response", "description": "Send to HITL for human review and editing", "output_field": "human_response"}, "determine_review_mode": {"action": "conditional", "config": {"condition": "(input_data.review_mode == 'hitl' OR review_mode == 'hitl') OR (input_data.require_human_review == true OR require_human_review == true)", "else_step": "auto_eval_content", "then_step": "prepare_hitl_review"}, "description": "Check whether to use HITL or auto-eval"}, "process_human_response": {"action": "process_human_input_response", "config": {"extract_edits": true}, "next_step": "finalize_hitl_result", "description": "Process the human reviewer's response", "output_field": "processed_response"}, "update_component_status": {"action": "update_page_components_status", "config": {"status": "approved", "page_from": "current_page", "reviewed_at_field": "review_result.reviewed_at", "reviewed_by_field": "review_result.reviewed_by"}, "next_step": "complete", "description": "Update page_components build_status and review info", "output_field": "status_updated"}, "finalize_escalation_result": {"action": "build_review_result", "config": {"review_mode": "escalated", "auto_eval_issues": "eval_result.issues"}, "next_step": "update_component_status", "description": "Build result from escalated HITL review", "output_field": "review_result"}, "process_escalation_response": {"action": "process_human_input_response", "config": {"extract_edits": true}, "next_step": "finalize_escalation_result", "description": "Process response from escalated review", "output_field": "escalation_processed"}}, "start_step": "determine_review_mode"}, "processing_mode": "task", "timeout_seconds": 600}
is_active              | t
created_at             | 2025-12-22 17:47:42.958031+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["content-review", "quality-assurance", "hitl", "auto-eval"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | active
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"review_mode": "string (optional) - 'hitl' for human review, 'auto' for LLM review. Defaults to auto.", "current_page": "object with name, title - the page being reviewed", "page_content": "object with sections[] containing rendered content", "reviewed_brief": "object with company info for context", "require_human_review": "boolean (optional) - if true, forces HITL mode. Defaults to false."}, "defaults": {"review_mode": "auto", "require_human_review": false}, "optional": ["reviewed_brief", "review_mode", "require_human_review"], "required": ["current_page", "page_content"]}
output_contract        | {"produces": {"edits": "object - any modifications made", "issues": "array - issues found (if not approved)", "content": "object - final content (possibly edited)", "approved": "boolean - whether content passed review", "review_mode": "string - hitl or auto-eval", "reviewed_at": "timestamp", "reviewed_by": "string - user email or eval-agent"}}
-[ RECORD 15 ]---+--
id                     | bfebf7c8-d181-4b76-b987-26f10e52b916
type                   | website-builder
display_name           | Website Builder
description            | Orchestrates complete website creation
category               | code-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Website build complete"}, "deploy_site": {"action": "call_agent", "config": {"target_role": "deployer", "input_fields": ["site_files", "input_data"], "timeout_seconds": 180}, "next_step": "complete", "description": "Deploy site to hosting", "output_field": "deployment_result"}, "develop_site": {"action": "call_agent", "config": {"target_role": "developer", "input_fields": ["input_data", "site_architecture", "site_content", "domain_analysis"], "timeout_seconds": 300}, "next_step": "wrap_multipage", "description": "Develop the HTML/CSS for the site", "output_field": "final_html"}, "spawn_wrapper": {"action": "spawn_agent", "config": {"role": "wrapper", "agent_type": "multipage-wrapper"}, "next_step": "spawn_deployer", "description": "Spawn multipage wrapper agent", "output_field": "spawned_wrapper"}, "analyze_domain": {"action": "call_agent", "config": {"target_role": "analyst", "input_fields": ["input_data"], "timeout_seconds": 120}, "next_step": "design_architecture", "description": "Analyze the domain name and objective", "output_field": "domain_analysis"}, "create_content": {"action": "call_agent", "config": {"target_role": "content", "input_fields": ["input_data", "domain_analysis", "site_architecture"], "timeout_seconds": 300}, "next_step": "develop_site", "description": "Create content for all site sections", "output_field": "site_content"}, "spawn_deployer": {"action": "spawn_agent", "config": {"role": "deployer", "agent_type": "site-deployer"}, "next_step": "analyze_domain", "description": "Spawn site deployer agent", "output_field": "spawned_deployer"}, "wrap_multipage": {"action": "call_agent", "config": {"target_role": "wrapper", "input_fields": ["final_html", "input_data"], "timeout_seconds": 60, "index_html_field": "final_html.final_html"}, "next_step": "deploy_site", "description": "Create about and contact pages, package as files map", "output_field": "site_files"}, "design_architecture": {"action": "call_agent", "config": {"target_role": "architect", "input_fields": ["input_data", "domain_analysis"], "timeout_seconds": 180}, "next_step": "create_content", "description": "Design the site structure and components", "output_field": "site_architecture"}, "spawn_domain_analyst": {"action": "spawn_agent", "config": {"role": "analyst", "agent_type": "domain-analyst"}, "next_step": "spawn_site_architect", "description": "Spawn domain analyst agent", "output_field": "spawned_analyst"}, "spawn_html_developer": {"action": "spawn_agent", "config": {"role": "developer", "agent_type": "html-developer"}, "next_step": "spawn_wrapper", "description": "Spawn HTML developer agent", "output_field": "spawned_developer"}, "spawn_site_architect": {"action": "spawn_agent", "config": {"role": "architect", "agent_type": "site-architect"}, "next_step": "spawn_content_creator", "description": "Spawn site architect agent", "output_field": "spawned_architect"}, "spawn_content_creator": {"action": "spawn_agent", "config": {"role": "content", "agent_type": "content-creator"}, "next_step": "spawn_html_developer", "description": "Spawn content creator agent", "output_field": "spawned_content"}}, "start_step": "spawn_domain_analyst"}, "ai_service": {"model": "claude-3-5-sonnet", "provider": "anthropic"}}
is_active              | t
created_at             | 2025-08-10 16:00:36.520323+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | []
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"dlq": "system.agent.{type}.dlq", "errors": "system.agent.{type}.errors", "requests": "system.agent.{type}.requests", "responses": "system.agent.{type}.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | {"steps": {"complete": {"action": "complete_workflow"}, "develop_site": {"action": "call_agent", "config": {"agent_type": "html-developer"}, "next_step": "complete"}, "analyze_domain": {"action": "call_agent", "config": {"agent_type": "domain-analyst"}, "next_step": "design_architecture"}, "create_content": {"action": "call_agent", "config": {"agent_type": "content-creator"}, "next_step": "develop_site"}, "design_architecture": {"action": "call_agent", "config": {"agent_type": "site-architect"}, "next_step": "create_content"}}, "start_step": "analyze_domain"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": false}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {"sections": [{"name": "company_info", "title": "Company Information", "questions": [{"type": "text", "field": "company_name", "label": "Company/Brand Name", "required": true}, {"type": "text", "field": "tagline", "label": "Tagline or Slogan", "required": false, "placeholder": "Your memorable one-liner"}, {"type": "textarea", "field": "about_us", "label": "About Your Company", "required": true, "placeholder": "Tell us about your company, mission, and what you do"}, {"type": "text", "field": "industry_type", "label": "Industry", "required": false, "placeholder": "e.g., Technology, Healthcare, Consulting"}]}, {"name": "offerings", "title": "Services & Products", "questions": [{"type": "json_array", "field": "services", "label": "Services or Products (list of {name, description})", "required": true, "placeholder": "[{\"name\": \"Service Name\", \"description\": \"What it does\"}]"}, {"type": "textarea", "field": "key_differentiators", "label": "What Makes You Different?", "required": true, "placeholder": "Your unique value propositions"}]}, {"name": "team", "title": "Team & Leadership", "questions": [{"type": "json_array", "field": "leadership_team", "label": "Leadership Team (list of {name, title, bio})", "required": false, "placeholder": "[{\"name\": \"Jane Doe\", \"title\": \"CEO\", \"bio\": \"Background and expertise\"}]"}]}, {"name": "portfolio", "title": "Portfolio & Social Proof", "questions": [{"type": "json_array", "field": "case_studies", "label": "Case Studies or Projects (list of {client, challenge, result})", "required": false, "placeholder": "[{\"client\": \"Company Name\", \"challenge\": \"Problem solved\", \"result\": \"Outcomes achieved\"}]"}, {"type": "text", "field": "client_count", "label": "Number of Clients/Customers", "required": false}]}, {"name": "contact", "title": "Contact Information", "questions": [{"type": "text", "field": "contact_email", "label": "Contact Email", "required": true}, {"type": "text", "field": "contact_phone", "label": "Phone Number", "required": false}, {"type": "text", "field": "headquarters", "label": "Location/Headquarters", "required": false, "placeholder": "City, Country"}]}, {"name": "design", "title": "Design Preferences", "questions": [{"type": "select", "field": "tone", "label": "Brand Tone", "default": "professional", "options": ["professional", "friendly", "bold", "innovative", "traditional", "playful"]}, {"type": "text", "field": "color_scheme", "label": "Preferred Color Scheme", "required": false, "placeholder": "e.g., Blue and white, Modern neutrals"}]}, {"name": "features", "title": "Website Features", "questions": [{"type": "boolean", "field": "has_blog", "label": "Include Blog/Insights Section?", "default": false}, {"type": "boolean", "field": "has_careers", "label": "Include Careers Page?", "default": false}, {"type": "text", "field": "primary_cta", "label": "Primary Call-to-Action", "required": false, "placeholder": "e.g., Contact Us, Get Started, Learn More"}, {"type": "text", "field": "primary_cta_url", "label": "Primary CTA Link", "required": false, "placeholder": "/contact or external URL"}]}, {"name": "audience", "title": "Target Audience", "questions": [{"type": "textarea", "field": "target_audience", "label": "Who is your target audience?", "required": true, "placeholder": "Describe your ideal visitors/customers"}]}]}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 16 ]---+--
id                     | d9d58405-8055-42c5-bf66-5263d57afbf5
type                   | content-site-builder
display_name           | Content Site Builder
description            | Orchestrates the complete content/publishing site build workflow
category               | orchestrator
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Content site build complete"}, "call_writer": {"action": "call_agent", "config": {"agent_type": "content-writer", "target_role": "writer", "input_fields": ["template_data", "build_plan", "brief_data", "input_data"], "timeout_seconds": 300}, "next_step": "call_assembler", "output_field": "content_data"}, "call_wrapper": {"action": "call_agent", "config": {"agent_type": "multipage-wrapper", "target_role": "wrapper", "input_fields": ["final_html", "input_data"], "timeout_seconds": 60}, "next_step": "call_deployer", "description": "Create about and contact pages, package as files map", "output_field": "site_files"}, "spawn_writer": {"action": "spawn_agent", "config": {"role": "writer", "agent_type": "content-writer"}, "next_step": "spawn_assembler"}, "call_deployer": {"action": "call_agent", "config": {"agent_type": "site-deployer", "target_role": "deployer", "input_fields": ["site_files", "input_data"], "timeout_seconds": 180}, "next_step": "complete", "output_field": "deployment_result"}, "spawn_wrapper": {"action": "spawn_agent", "config": {"role": "wrapper", "agent_type": "multipage-wrapper"}, "next_step": "spawn_deployer"}, "call_architect": {"action": "call_agent", "config": {"agent_type": "content-site-architect", "target_role": "architect", "input_fields": ["build_plan", "brief_data", "input_data"], "timeout_seconds": 120}, "next_step": "call_writer", "output_field": "template_data"}, "call_assembler": {"action": "call_agent", "config": {"agent_type": "html-assembler", "target_role": "assembler", "input_fields": ["content_data", "template_data", "brief_data", "input_data"], "timeout_seconds": 120}, "next_step": "call_wrapper", "output_field": "final_html"}, "spawn_deployer": {"action": "spawn_agent", "config": {"role": "deployer", "agent_type": "site-deployer"}, "next_step": "call_strategist"}, "call_strategist": {"action": "call_agent", "config": {"agent_type": "site-strategist", "target_role": "strategist", "input_fields": ["input_data", "brief_data"], "timeout_seconds": 120}, "next_step": "call_architect", "output_field": "build_plan"}, "spawn_architect": {"action": "spawn_agent", "config": {"role": "architect", "agent_type": "content-site-architect"}, "next_step": "spawn_writer"}, "spawn_assembler": {"action": "spawn_agent", "config": {"role": "assembler", "agent_type": "html-assembler"}, "next_step": "spawn_wrapper"}, "spawn_strategist": {"action": "spawn_agent", "config": {"role": "strategist", "agent_type": "site-strategist"}, "next_step": "spawn_architect"}}, "start_step": "spawn_strategist"}, "processing_mode": "orchestration", "timeout_seconds": 900}
is_active              | t
created_at             | 2025-12-04 08:42:32.269013+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["orchestration", "site-building", "content-site"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {"sections": [{"name": "publication", "title": "Publication Identity", "questions": [{"type": "text", "field": "publication_name", "label": "Publication/Site Name", "required": true}, {"type": "text", "field": "tagline", "label": "Tagline", "required": false}, {"type": "select", "field": "editorial_tone", "label": "Editorial Tone", "default": "magazine_polished", "options": ["news_formal", "magazine_polished", "blog_casual", "technical"]}]}, {"name": "content_structure", "title": "Content Structure", "questions": [{"type": "textarea", "field": "categories", "label": "Content Categories (one per line)", "required": true}, {"type": "select", "field": "publishing_frequency", "label": "Publishing Frequency", "default": "weekly", "options": ["daily", "weekly", "occasional"]}]}, {"name": "monetization", "title": "Monetization", "questions": [{"type": "select", "field": "monetization_model", "label": "Revenue Model", "default": "advertising", "options": ["advertising", "subscription", "affiliate", "none"]}, {"type": "boolean", "field": "newsletter_signup", "label": "Include Newsletter Signup?", "default": true}]}]}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 17 ]---+--
id                     | cb3a90d6-85a8-4650-b3a7-1943df9d0714
type                   | content-creator-hero-without-research
display_name           | Content Creator (Hero - No Research)
description            | Generates hero sections for websites without performing research; uses direct input only.
category               | adapter
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return hero content"}, "generate_hero_content": {"action": "execute_llm_prompt", "config": {"input_fields": ["input_data"], "prompt_template": "Write a compelling hero section for {{.business_name}}, a {{.business_type}}. Include a powerful headline and an engaging subheadline that captures attention and communicates the core value proposition."}, "next_step": "complete", "description": "Generate hero section"}}, "start_step": "generate_hero_content"}, "ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "max_tokens": 1500, "temperature": 0.7, "processing_mode": "task"}
is_active              | t
created_at             | 2025-10-30 14:25:32.669321+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["content-creation", "text-generation"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 18 ]---+--
id                     | 7edb4863-c662-49b6-aa4b-7609f9976522
type                   | site-architect
display_name           | Site Architect
description            | Plans website structure and navigation
category               | data-driven
default_config         | {"workflow": {"steps": {"design": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["input_data", "domain_analysis", "reviewed_brief", "component_library"], "output_format": "json", "prompt_template": "You are a site architect designing a professional website.\n\n## Business Information\nDomain: {{.input_data.domain}}\nCompany: {{if .reviewed_brief.company_name}}{{.reviewed_brief.company_name}}{{else}}{{.input_data.domain}}{{end}}\nIndustry: {{.reviewed_brief.industry}}\nTone: {{.reviewed_brief.tone}}\nTagline: {{.reviewed_brief.tagline}}\n\n## Services\n{{range .reviewed_brief.services}}- {{.name}}: {{.description}}\n{{end}}\n\n## Available Components\nYou MUST select sections from this list. Use the function name exactly as shown:\n\n{{.component_library.for_prompt}}\n\n## Design Task\nCreate a site architecture with pages appropriate for this business.\n\n## Rules\n1. Every page MUST have at least 2 sections\n2. Use ONLY function names from the available components list above\n3. The index page should have 3-5 sections (hero, then supporting sections)\n4. Secondary pages should have 2-4 sections\n5. Common patterns:\n   - Landing pages: hero → features → social_proof → call_to_action\n   - About pages: hero → generic-text-block → call_to_action\n   - Services pages: hero → features → call_to_action\n   - Contact pages: hero → generic-text-block → call_to_action\n\n## Output JSON Format\n```json\n{\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Company Name\",\n      \"nav_label\": \"Home\",\n      \"nav_order\": 1,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"page_type\": \"landing\",\n      \"sections\": [\"hero\", \"features\", \"social_proof\", \"call_to_action\"]\n    },\n    {\n      \"name\": \"about\",\n      \"title\": \"About Us | Company Name\",\n      \"nav_label\": \"About\",\n      \"nav_order\": 2,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"page_type\": \"content\",\n      \"sections\": [\"hero\", \"generic-text-block\", \"call_to_action\"]\n    }\n  ],\n  \"style_collection\": \"default\",\n  \"needs_logo\": true,\n  \"needs_images\": true,\n  \"image_prompts\": {\n    \"hero\": \"Professional hero image for [describe appropriate imagery]\",\n    \"logo\": \"Modern logo for [company name]\"\n  }\n}\n```\n\nReturn ONLY valid JSON, no markdown backticks or explanation."}, "next_step": "complete", "description": "Design site structure using available components", "output_field": "site_architecture"}, "complete": {"action": "complete_workflow", "config": {"output": {"pages": "site_architecture.pages", "needs_logo": "site_architecture.needs_logo", "needs_images": "site_architecture.needs_images", "image_prompts": "site_architecture.image_prompts", "style_collection": "site_architecture.style_collection"}}, "description": "Return site architecture"}, "load_components": {"action": "load_component_library", "config": {"component_level": "section", "format_for_prompt": true}, "next_step": "design", "description": "Load available section components from database", "output_field": "component_library"}}, "start_step": "load_components"}}
is_active              | t
created_at             | 2025-08-10 16:00:36.520323+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["planning", "structure", "navigation"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | {./agent-chassis,-config,configs/agent-chassis.yaml}
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"dlq": "system.agent.site-architect.dlq", "errors": "system.agent.site-architect.errors", "requests": "system.agent.site-architect.requests", "responses": "system.agent.site-architect.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | {"steps": {"complete": {"action": "complete_workflow", "description": "Return site architecture"}, "design_architecture": {"action": "execute_llm_prompt", "config": {"prompt_template": "Design the site architecture for: {{.business_type}}. Include: navigation structure, page hierarchy, user flows, and key components."}, "next_step": "complete", "description": "Design site structure and navigation"}}, "start_step": "design_architecture"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": false}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"domain_analysis": "object"}, "required": ["input_data", "domain_analysis"]}
output_contract        | {"format": {"type": "object", "description": "Site structure, pages, and component layout"}, "produces": "site_architecture"}
-[ RECORD 19 ]---+--
id                     | c6432aa9-4ec3-418f-ae9a-d51df4dc627d
type                   | content-creator-cta
display_name           | Call-to-Action Writer
description            | Specialized in writing persuasive call-to-action sections with urgency and clear value
category               | data-driven
default_config         | {"model": "claude-3-5-sonnet", "workflow": {"steps": {"complete": {"action": "complete_workflow"}, "generate_content": {"action": "execute_llm_prompt", "config": {"input_fields": ["input_data"]}, "next_step": "complete"}}, "start_step": "generate_content"}, "ai_service": {"model": "claude-3-5-sonnet", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "max_tokens": 2000, "temperature": 0.8, "processing_mode": "task", "prompt_template": "Write a strong call-to-action section for {{.business_name}}. Include compelling action text and a clear reason to act now. Make it urgent, persuasive, and action-oriented."}
is_active              | t
created_at             | 2025-10-14 19:00:09.810508+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["content", "cta", "conversion", "persuasion"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.content-creator-cta", "process": "system.agent.content-creator-cta.process", "response": "system.responses.content-creator-cta"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 20 ]---+--
id                     | 3901d4e2-dadd-4e9b-94cf-d5e2001b4acc
type                   | content-creator
display_name           | Content Creator V2
description            | Creates content for component-based pages (v2 - receives current_page object)
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return content"}, "create_content": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "output_type": "json", "input_fields": ["input_data", "current_page", "page_plan"], "prompt_template": "You are a professional copywriter creating content for a website page.\n\nCONTEXT:\nDomain: {{.input_data.domain}}\nObjective: {{.input_data.objective}}\nMarketing Model: {{.input_data.model}}\n\nPAGE TO CREATE:\nName: {{.current_page.name}}\nTitle: {{.current_page.title}}\nPurpose: {{.current_page.purpose}}\nComponents: {{.current_page.components}}\n\nCreate content for EACH component listed above.\n\nComponent guidelines:\n- hero-*: headline (5-8 words), subheadline (15-25 words), cta_text\n- services-*: items array with name, description, icon_suggestion\n- features-*: items array with name, benefit, detail\n- testimonials-*: items array with quote, name, title, company\n- team-*: items array with name, role, bio\n- pricing-*: tiers array with name, price, features, cta\n- faq-*: items array with question, answer\n- cta-*: headline, supporting_text, button_text\n- contact-*: intro, email, phone, address\n- about-*: paragraphs array, key_points\n- footer-*: company_name, tagline, link_groups, copyright\n\nOUTPUT FORMAT (valid JSON):\n{\n  \"page_name\": \"index\",\n  \"hero\": {\"headline\": \"...\", \"subheadline\": \"...\", \"cta_text\": \"...\", \"cta_url\": \"#\"},\n  \"sections\": [\n    {\"type\": \"component-type\", \"content\": {...}}\n  ],\n  \"meta\": {\"page_title\": \"...\", \"meta_description\": \"...\"},\n  \"footer\": {\"company_name\": \"...\", \"tagline\": \"...\", \"copyright\": \"...\"}\n}"}, "next_step": "complete", "description": "Create content for page components", "output_field": "content_result"}}, "start_step": "create_content"}, "processing_mode": "task", "timeout_seconds": 180}
is_active              | t
created_at             | 2025-12-18 13:07:22.761389+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["content-generation", "writing", "copywriting", "seo", "memory-enabled", "style-adaptive"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "1000m", "memory": "2Gi"}, "requests": {"cpu": "200m", "memory": "512Mi"}}
topics                 | {"dlq": "system.agent.content-creator.dlq", "errors": "system.agent.content-creator.errors", "requests": "system.agent.content-creator.requests", "responses": "system.agent.content-creator.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 2
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"current_page": "object with name, components"}, "required": ["input_data", "current_page"]}
output_contract        | {"format": {"type": "object", "properties": {"hero": "object", "sections": "array"}}, "produces": "page_content"}
-[ RECORD 21 ]---+--
id                     | 411211f5-7da2-4320-8f6f-0194ea23848c
type                   | simple-content-writer-with-approval
display_name           | Simple Content Writer with HITL Approval
description            | Generates content for organisations and waits for human approval before completion
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return approved content with metadata"}, "generate_draft": {"action": "execute_llm_prompt", "config": {"input_fields": ["input_data"], "prompt_template": "Write a brief, professional description for {{.input_data.business_name}}. Context: {{.input_data.business_description}}. Focus on how they help organisations with {{.input_data.business_type}}. Keep it to 3-4 sentences that capture the essence of their value proposition."}, "next_step": "await_human_approval", "description": "Generate initial content draft"}, "process_approval": {"action": "process_data", "config": {"input_fields": ["generate_draft", "await_human_approval"], "output_format": {"content": "{{.generate_draft.result}}", "approved_at": "{{.await_human_approval.timestamp}}", "approved_by": "{{.await_human_approval.approved_by}}", "approval_status": "{{.await_human_approval.approved}}", "approval_comments": "{{.await_human_approval.comments}}"}}, "next_step": "complete", "description": "Process approval response and prepare final output"}, "await_human_approval": {"action": "await_approval", "config": {"timeout_seconds": 300, "notification_data": {"type": "content_approval", "title": "Content Approval Required for {{.input_data.business_name}}", "metadata": {"business_name": "{{.input_data.business_name}}", "business_type": "{{.input_data.business_type}}"}, "description": "Please review and approve the generated organisational description", "content_field": "generate_draft"}, "include_generated_content": true}, "next_step": "process_approval", "description": "Wait for human approval of generated content"}}, "start_step": "generate_draft"}, "ai_service": {"model": "claude-3-5-sonnet", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "max_tokens": 500, "temperature": 0.7, "processing_mode": "task"}
is_active              | t
created_at             | 2025-11-03 14:57:40.772284+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["content-generation", "human-approval", "hitl", "organisational"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"dlq": "system.agent.simple-content-writer-with-approval.dlq", "errors": "system.agent.simple-content-writer-with-approval.errors", "requests": "system.agent.simple-content-writer-with-approval.requests", "responses": "system.agent.simple-content-writer-with-approval.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | [{"name": "ANTHROPIC_API_KEY", "valueFrom": {"secretKeyRef": {"key": "api-key", "name": "anthropic-api-key"}}}]
version                | 1
previous_version_id    | 
task_workflow          | {"steps": {"complete": {"action": "complete_workflow", "description": "Return approved content"}, "generate_draft": {"action": "execute_llm_prompt", "config": {"input_fields": ["input_data"], "prompt_template": "Write a brief, professional description for {{.input_data.business_name}}. Context: {{.input_data.business_description}}. Focus on how they help organisations with {{.input_data.business_type}}. Keep it to 3-4 sentences that capture the essence of their value proposition."}, "next_step": "await_human_approval", "description": "Generate initial content draft"}, "process_approval": {"action": "process_data", "config": {"input_fields": ["generate_draft", "await_human_approval"]}, "next_step": "complete", "description": "Process approval response"}, "await_human_approval": {"action": "await_approval", "config": {"timeout_seconds": 300, "notification_data": {"type": "content_approval", "title": "Content Approval Required", "description": "Please review and approve the generated organisational description"}, "include_generated_content": true}, "next_step": "process_approval", "description": "Wait for human approval"}}, "start_step": "generate_draft"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": false}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 22 ]---+--
id                     | 39ab6388-d7b8-4286-8bff-9a80cf18ca63
type                   | content-creator-testimonials
display_name           | Testimonials Writer
description            | Specialized in writing authentic, emotionally resonant customer testimonials
category               | data-driven
default_config         | {"model": "claude-3-5-sonnet", "workflow": {"steps": {"complete": {"action": "complete_workflow"}, "generate_content": {"action": "execute_llm_prompt", "config": {"input_fields": ["input_data"]}, "next_step": "complete"}}, "start_step": "generate_content"}, "ai_service": {"model": "claude-3-5-sonnet", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "max_tokens": 2000, "temperature": 0.7, "processing_mode": "task", "prompt_template": "Create 2-3 realistic customer testimonials for {{.input_data.business_name}}, a {{.business_type}}. Make them authentic, specific, and emotionally resonant. Include customer names, contexts, and specific benefits they experienced."}
is_active              | t
created_at             | 2025-10-14 19:00:09.814169+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["content", "testimonials", "social-proof", "trust"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.content-creator-testimonials", "process": "system.agent.content-creator-testimonials.process", "response": "system.responses.content-creator-testimonials"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 23 ]---+--
id                     | 1db46661-2811-4aca-8a59-5212e36d3b88
type                   | image-generator
display_name           | Image Generator
description            | Creates images using AI generation with S3 storage (orchestrator mode)
category               | adapter
default_config         | {"model": "sdxl", "provider": "stability_ai", "workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return image URI"}, "generate": {"action": "generate_image", "next_step": "complete", "description": "Generate image and upload to S3"}}, "start_step": "generate"}, "api_config": {"base_url": "https://api.stability.ai/v1/generation/stable-diffusion-xl-1024-v1-0/text-to-image", "api_key_env_var": "IMAGE_API_KEY", "timeout_seconds": 60}, "capabilities": {"storage": {"enabled": true, "provider": "s3", "bucket_env_var": "IMAGE_BUCKET", "region_env_var": "S3_REGION", "endpoint_env_var": "S3_ENDPOINT", "access_key_env_var": "AWS_ACCESS_KEY_ID", "secret_key_env_var": "AWS_SECRET_ACCESS_KEY"}}, "image_settings": {"max_width": 2048, "max_height": 2048, "default_width": 1920, "default_format": "png", "default_height": 1080}, "processing_mode": "adapter"}
is_active              | t
created_at             | 2025-10-27 19:42:15.923354+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["image-generation", "text-to-image", "s3-storage", "orchestration"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"dlq": "system.agent.image-generator.dlq", "errors": "system.agent.image-generator.errors", "requests": "system.agent.image-generator.requests", "responses": "system.agent.image-generator.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | [{"name": "AWS_ACCESS_KEY_ID", "valueFrom": {"secretKeyRef": {"key": "AWS_ACCESS_KEY_ID", "name": "personae-storage-secrets"}}}, {"name": "AWS_SECRET_ACCESS_KEY", "valueFrom": {"secretKeyRef": {"key": "AWS_SECRET_ACCESS_KEY", "name": "personae-storage-secrets"}}}, {"name": "S3_ENDPOINT", "valueFrom": {"configMapKeyRef": {"key": "S3-ENDPOINT", "name": "storage-config"}}}, {"name": "S3_REGION", "valueFrom": {"configMapKeyRef": {"key": "S3-REGION", "name": "storage-config"}}}, {"name": "IMAGE_BUCKET", "valueFrom": {"configMapKeyRef": {"key": "image_bucket", "name": "storage-config"}}}, {"name": "IMAGE_API_KEY", "valueFrom": {"secretKeyRef": {"key": "api-key", "name": "image-api-credentials"}}}]
version                | 2
previous_version_id    | 
task_workflow          | {"steps": {"complete": {"action": "complete_workflow", "description": "Return image URI"}, "generate_and_upload": {"action": "generate_image_and_upload", "next_step": "complete", "description": "Generate image and upload to S3"}}, "start_step": "generate_and_upload"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 24 ]---+--
id                     | c16638a7-c867-481e-a71e-6ae751ac961b
type                   | site-strategist
display_name           | Site Strategist
description            | Creates strategic build plans using behavioral psychology and briefing data. Works with intake orchestrator flow.
category               | strategy
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return the Build Plan"}, "generate_build_plan": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "max_tokens": 10000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["input_data", "brief_data"], "output_field": "build_plan_json", "prompt_template": "You are an english website strategist creating a Build Plan based on behavioural psychology.\n\nWebsite Request:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Model: {{.input_data.model}}\n\n{{if .brief_data}}Brief Data:\n{{.brief_data}}\n{{end}}\n\nBehavioural Models:\n- AIDA: Attention → Interest → Desire → (Conviction) -> Action\n- PAS: Problem → Agitate → Solution\n- FAB: Features → Advantages → Benefits\n- 4Ps: Promise → Picture → Proof → Push\n\nAvailable Components: header, hero, features, social_proof, pricing, faq, call_to_action, footer\n\nCreate a build plan that maps the behavioural model to sections with specific guidance.\n\nReturn ONLY valid JSON:\n{\n  \"model\": \"AIDA\",\n  \"sections\": [\"header\", \"hero\", \"features\", \"social_proof\", \"pricing\", \"faq\", \"call_to_action\", \"footer\"],\n  \"section_guidance\": {\n    \"hero\": {\n      \"stage\": \"Attention\",\n      \"purpose\": \"Grab attention with bold value proposition\",\n      \"key_message\": \"Main benefit headline\",\n      \"tone\": \"Confident, clear\"\n    },\n    \"features\": {\n      \"stage\": \"Interest\",\n      \"purpose\": \"Build interest with capabilities\",\n      \"key_message\": \"What it does\",\n      \"tone\": \"Informative\"\n    }\n  },\n  \"theme\": \"tech-saas\"\n}\n\nProvide section_guidance for each section in the sections array. Keep this guidance concise. Return ONLY the JSON object."}, "next_step": "complete"}}, "start_step": "generate_build_plan"}, "processing_mode": "task", "timeout_seconds": 120}
is_active              | t
created_at             | 2025-11-29 17:24:43.168225+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["strategy", "planning", "llm"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | strategist
status                 | active
domain_tags            | ["website"]
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 25 ]---+--
id                     | 70c139b3-f344-41a7-a8a1-d720bedba4ed
type                   | portfolio-architect
display_name           | Portfolio Site Architect
description            | Assembles portfolio/showcase sites with galleries, case studies, and visual layouts
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return the portfolio site template"}, "assemble_portfolio_site": {"action": "assemble_from_library", "config": {"site_type": "portfolio", "input_fields": ["build_plan_data", "brief_data"], "default_sections": ["header", "hero_visual", "work_grid", "case_study", "client_logos", "about", "contact", "footer"], "component_category": "portfolio-site"}, "next_step": "complete", "description": "Assemble portfolio site template from component library"}}, "start_step": "assemble_portfolio_site"}, "processing_mode": "task", "timeout_seconds": 180}
is_active              | t
created_at             | 2025-11-28 13:10:10.997945+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["build", "assemble", "database", "portfolio-site", "showcase"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 26 ]---+--
id                     | 5946a27b-38ab-41e8-8b49-7bc1a4b626b8
type                   | page-content-writer
display_name           | Page Content Writer
description            | Writes content for a single page. Loads section components, renders templates with brief data, calls LLM for content sections needing original writing. Can spawn research-agent for research-backed content.
category               | specialist
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_field": "page_content"}}, "compile_page": {"action": "compile_page_sections", "config": {"page_from": "input_data.current_page", "inject_footer": true, "inject_header": true, "sections_from": "processed_sections", "include_research_ids": true}, "next_step": "complete", "description": "Compile all sections into page output", "output_field": "page_content"}, "build_render_context": {"action": "build_render_context", "config": {"sources": {"page": "input_data.current_page", "site": "input_data.site_record", "brief": "input_data.reviewed_brief", "style": "input_data.style_collection", "assets": "brand_assets"}}, "next_step": "process_sections_loop", "description": "Build render context from brief, site, and brand data", "output_field": "render_context"}, "load_page_components": {"action": "load_page_section_components", "config": {"page_from": "input_data.current_page", "sections_from": "input_data.current_page.sections", "include_templates": true, "include_input_schema": true}, "next_step": "build_render_context", "description": "Load component definitions for this page's sections", "output_field": "section_components"}, "spawn_research_agent": {"action": "spawn_agent", "config": {"role": "researcher", "agent_type": "research-agent", "await_response": true}, "next_step": "load_page_components", "description": "Spawn research agent in case sections need research", "output_field": "researcher_info"}, "process_sections_loop": {"action": "loop", "config": {"loop_var": "current_section", "iterate_over": "section_components.components", "sub_workflow": {"steps": {"render_section": {"action": "render_component", "config": {"output_html": true, "content_from": "generated_content.result", "context_from": "render_context", "component_from": "current_section"}, "description": "Render LLM-generated content into component template", "output_field": "section_output"}, "call_researcher": {"action": "call_agent", "config": {"agent_type": "research-agent", "target_role": "researcher", "input_fields": ["current_section", "reviewed_brief", "site_record"], "timeout_seconds": 90}, "next_step": "generate_content", "description": "Research topic for this section", "output_field": "research_result"}, "generate_content": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5", "provider": "anthropic", "max_tokens": 2000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "input_fields": ["current_section", "render_context", "reviewed_brief", "current_page"], "output_format": "json", "prompt_template": "Write content for the {{.current_section.function}} section of {{.current_page.title}}.\n\n## Company Context\nCompany: {{.render_context.company_name}}\nIndustry: {{.render_context.industry}}\nTone: {{.render_context.tone}}\nTarget Audience: {{.render_context.target_audience}}\nServices: {{.reviewed_brief.services}}\nTagline: {{.render_context.tagline}}\n\n## Section Requirements\nComponent: {{.current_section.name}}\nFunction: {{.current_section.function}}\nPurpose: {{.current_section.description}}\n\n## Data Schema Required\n{{.current_section.input_schema}}\n\n{{if .research_result}}\n## Research Findings\n{{.research_result.response.summary}}\n\nSources:\n{{range $index, $src := .research_result.response.sources}}\n- [{{$index}}] {{$src.title}} ({{$src.domain}})\n{{end}}\n{{end}}\n\n## Task\nWrite compelling, specific content for this section.\n\nReturn JSON with these EXACT field names (use the ones that apply to this component type):\n\n### For Hero/Banner sections:\n```json\n{\n  \"headline\": \"Your Compelling Main Headline\",\n  \"subheadline\": \"Supporting text that expands on the headline\",\n  \"primary_cta\": \"Get Started\",\n  \"primary_cta_url\": \"/contact.html\",\n  \"secondary_cta\": \"Learn More\",\n  \"secondary_cta_url\": \"/about.html\"\n}\n```\n\n### For Feature/Services sections:\n```json\n{\n  \"headline\": \"Section Headline\",\n  \"subheadline\": \"Brief introduction\",\n  \"features\": [\n    {\"name\": \"Feature Name\", \"description\": \"Feature description\", \"icon\": \"icon-name\"},\n    {\"name\": \"Feature 2\", \"description\": \"Description 2\", \"icon\": \"icon-name\"}\n  ]\n}\n```\n\n### For CTA/Call-to-Action sections:\n```json\n{\n  \"headline\": \"Ready to Get Started?\",\n  \"subheadline\": \"Contact us today\",\n  \"primary_cta\": \"Contact Us\",\n  \"primary_cta_url\": \"/contact.html\"\n}\n```\n\n### For Text/Content sections:\n```json\n{\n  \"heading\": \"Section Heading\",\n  \"content\": \"Paragraph content here...\"\n}\n```\n\nRules:\n- Use the EXACT field names shown above (headline, subheadline, primary_cta, primary_cta_url, etc.)\n- No placeholder text like [Your Company] or Lorem ipsum\n- Be specific to this company and industry\n- Professional but engaging tone matching the brief\n- Include source citations [0], [1] if research was provided"}, "next_step": "render_section", "description": "Generate content for this section", "output_field": "generated_content"}, "check_render_mode": {"action": "conditional", "config": {"condition": "current_section.render_mode == 'agent' OR current_section.needs_llm == true", "else_step": "render_from_template", "then_step": "check_needs_research"}, "description": "Check if section needs LLM or just template"}, "check_needs_research": {"action": "conditional", "config": {"condition": "current_section.needs_research == true", "else_step": "generate_content", "then_step": "call_researcher"}, "description": "Check if section needs research first"}, "render_from_template": {"action": "render_component", "config": {"output_html": true, "content_from": "render_context", "context_from": "render_context", "component_from": "current_section"}, "description": "Render section from template with brief data only", "output_field": "section_output"}}, "start_step": "check_render_mode"}, "max_iterations": 15}, "next_step": "compile_page", "description": "Process each section - template render or LLM generate", "output_field": "processed_sections"}}, "start_step": "spawn_research_agent"}, "processing_mode": "task", "timeout_seconds": 180}
is_active              | t
created_at             | 2025-12-22 17:47:17.609605+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["content-generation", "template-rendering", "research"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | active
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"site_record": "object with site_id, domain", "brand_assets": "object with logo, images (optional)", "current_page": "object with name, title, sections[]", "reviewed_brief": "object with company_name, services, about_us, etc", "style_collection": "object with colors, component refs"}, "required": ["current_page", "site_record", "reviewed_brief"]}
output_contract        | {"produces": {"sections": "array of {component_id, rendered_html, content_data}", "page_name": "string", "research_ids": "array of research result UUIDs used"}}
-[ RECORD 27 ]---+--
id                     | a76658cc-f895-4869-8ce2-69efe88183ad
type                   | site-publisher
display_name           | Site Publisher
description            | Publishes websites to storage buckets
category               | adapter
default_config         | {"workflow": {"steps": {"publish": {"action": "upload_to_s3", "config": {"bucket": "websites", "public": true}, "next_step": "complete", "description": "Upload site files to hosting"}, "complete": {"action": "complete_workflow", "description": "Return published site URL"}}, "start_step": "publish"}, "local_actions": ["validate_input", "upload_to_s3", "complete_workflow"], "storage_config": {"bucket": "personae-prod-uk001-site-assets", "region": "us-east-005", "endpoint": "https://s3.us-east-005.backblazeb2.com", "provider": "s3", "use_path_style": true, "access_key_env_var": "AWS_ACCESS_KEY_ID", "secret_key_env_var": "AWS_SECRET_ACCESS_KEY"}, "enable_local_actions": true}
is_active              | t
created_at             | 2025-08-10 16:00:36.520323+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["deployment", "hosting", "publishing", "s3"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | {./agent-chassis,-config,configs/agent-chassis.yaml}
resources              | {"limits": {"cpu": "1000m", "memory": "2Gi"}, "requests": {"cpu": "200m", "memory": "512Mi"}}
topics                 | {"dlq": "system.agent.site-publisher.dlq", "errors": "system.agent.site-publisher.errors", "requests": "system.agent.site-publisher.requests", "responses": "system.agent.site-publisher.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | [{"name": "ENABLE_LOCAL_ACTIONS", "value": "true"}, {"name": "LOCAL_ACTION_MODULES", "value": "storage_actions"}]
version                | 1
previous_version_id    | 
task_workflow          | {"steps": {"publish": {"action": "upload_to_s3", "config": {"bucket": "websites", "public": true}, "next_step": "complete", "description": "Upload site files to hosting"}, "complete": {"action": "complete_workflow", "description": "Return published site URL"}}, "start_step": "publish"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": false}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 28 ]---+--
id                     | 43153077-5b48-4638-9b7c-0ae089ff50e0
type                   | content_researcher
display_name           | Content Researcher
description            | Researches content for websites
category               | code-driven
default_config         | {"workflow": {"steps": {"process": {"action": "execute_llm_prompt", "config": {"input_fields": ["input_data"]}, "next_step": "complete", "description": "Research topic and return findings"}, "complete": {"action": "complete_workflow", "description": "Return research findings"}}, "start_step": "process"}, "ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "max_tokens": 1500, "temperature": 0.7, "processing_mode": "task", "prompt_template": "Research background information about {{.business_type}} businesses. Focus on: industry trends, common customer pain points, and key value propositions. Provide 3-4 concise bullet points."}
is_active              | t
created_at             | 2025-08-21 11:48:56.386838+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | []
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"dlq": "system.agent.{type}.dlq", "errors": "system.agent.{type}.errors", "requests": "system.agent.{type}.requests", "responses": "system.agent.{type}.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | {"steps": {"execute": {"action": "execute_llm_prompt", "next_step": "complete", "description": "Execute task"}, "complete": {"action": "complete_workflow", "description": "Complete"}}, "start_step": "execute"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 29 ]---+--
id                     | db788512-6273-4f83-bda3-057edd3ba743
type                   | html-assembler
display_name           | HTML Assembler Agent
description            | Assembles final HTML from template and content, injects CSS themes and JS snippets.
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return assembled HTML"}, "assemble_html": {"action": "assemble_full_page", "config": {"minify": false, "inject_js": true, "inject_css": true, "theme_field": "content_data.content_json.result.theme", "content_field": "content_data.content_json", "template_field": "template_data.html_template", "theme_tags_field": "content_data.content_json.result.theme_tags"}, "next_step": "complete", "description": "Assemble HTML with CSS and JS", "output_field": "final_html"}}, "start_step": "assemble_html"}, "processing_mode": "task", "timeout_seconds": 120}
is_active              | t
created_at             | 2025-11-28 10:41:21.475847+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["html", "assembly", "css", "javascript"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | executor
status                 | active
domain_tags            | ["website"]
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 30 ]---+--
id                     | 6187d808-0d25-441b-b7b3-40af562878af
type                   | generic
display_name           | Generic Orchestrator
description            | Generic agent that can spawn groups and orchestrate workflows
category               | code-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Complete workflow"}, "spawn_adder": {"action": "spawn_agent", "config": {"role": "adder", "agent_type": "calculator"}, "next_step": "spawn_multiplier", "description": "Spawn addition calculator"}, "perform_addition": {"action": "call_agent", "config": {"agent_type": "calculator", "input_field": "first_calc", "target_role": "adder"}, "next_step": "perform_multiplication", "description": "Addition calculation"}, "spawn_multiplier": {"action": "spawn_agent", "config": {"role": "multiplier", "agent_type": "calculator"}, "next_step": "perform_addition", "description": "Spawn multiplication calculator"}, "aggregate_results": {"action": "aggregate_data", "config": {"strategy": "group_responses", "response_fields": ["perform_addition", "perform_multiplication"]}, "next_step": "complete", "description": "Aggregate calculation results"}, "perform_multiplication": {"action": "call_agent", "config": {"agent_type": "calculator", "input_field": "second_calc", "target_role": "multiplier"}, "next_step": "aggregate_results", "description": "Multiplication calculation"}}, "start_step": "spawn_adder"}}
is_active              | t
created_at             | 2025-08-20 10:26:26.116535+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["orchestration", "spawn_group"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"dlq": "system.agent.{type}.dlq", "errors": "system.agent.{type}.errors", "requests": "system.agent.{type}.requests", "responses": "system.agent.{type}.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | {"steps": {"complete": {"action": "complete_workflow", "description": "Complete the orchestration"}, "spawn_website_team": {"action": "spawn_group", "config": {"group_type": "website-builder"}, "next_step": "start_website_workflow", "description": "Spawn the website builder team"}, "start_website_workflow": {"action": "start_orchestration", "next_step": "complete", "description": "Start the website builder orchestration"}}, "start_step": "spawn_website_team"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": false}
agent_category         | coordinator
status                 | active
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 31 ]---+--
id                     | 59ca0d97-577f-45f8-963a-6c7ba2754da4
type                   | image-generator
display_name           | Image Generator
description            | Creates images using AI generation
category               | adapter
default_config         | {"model": "sdxl", "provider": "stability_ai", "workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return image URI"}, "generate": {"action": "generate_image", "next_step": "complete", "description": "Generate image and upload to S3"}}, "start_step": "generate"}, "capabilities": {"storage": {"enabled": true, "provider": "s3", "bucket_env_var": "ASSETS_BUCKET", "region_env_var": "S3_REGION", "endpoint_env_var": "S3_ENDPOINT", "access_key_env_var": "AWS_ACCESS_KEY_ID", "secret_key_env_var": "AWS_SECRET_ACCESS_KEY"}}, "processing_mode": "adapter"}
is_active              | t
created_at             | 2025-08-10 16:00:36.153034+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | []
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"dlq": "system.agent.{type}.dlq", "errors": "system.agent.{type}.errors", "requests": "system.agent.{type}.requests", "responses": "system.agent.{type}.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | [{"name": "AWS_ACCESS_KEY_ID", "valueFrom": {"secretKeyRef": {"key": "AWS_ACCESS_KEY_ID", "name": "personae-storage-secrets"}}}, {"name": "AWS_SECRET_ACCESS_KEY", "valueFrom": {"secretKeyRef": {"key": "AWS_SECRET_ACCESS_KEY", "name": "personae-storage-secrets"}}}, {"name": "S3_ENDPOINT", "valueFrom": {"configMapKeyRef": {"key": "S3-ENDPOINT", "name": "storage-config"}}}, {"name": "S3_REGION", "valueFrom": {"configMapKeyRef": {"key": "S3-REGION", "name": "storage-config"}}}, {"name": "IMAGE_BUCKET", "valueFrom": {"configMapKeyRef": {"key": "image_bucket", "name": "storage-config"}}}, {"name": "ASSETS_BUCKET", "valueFrom": {"configMapKeyRef": {"key": "assets_bucket", "name": "storage-config"}}}]
version                | 1
previous_version_id    | 
task_workflow          | {"steps": {"execute": {"action": "execute_llm_prompt", "next_step": "complete", "description": "Execute task"}, "complete": {"action": "complete_workflow", "description": "Complete"}}, "start_step": "execute"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 32 ]---+--
id                     | 86c7e418-3eed-41ef-b678-bc592eabc10c
type                   | landing-page-builder
display_name           | Landing Page Builder
description            | Orchestrates the complete landing page build workflow - spawns specialist agents and coordinates them to build conversion-focused landing pages
category               | orchestrator
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Landing page build complete"}, "call_writer": {"action": "call_agent", "config": {"agent_type": "content-writer", "target_role": "writer", "input_fields": ["template_data", "build_plan", "brief_data", "input_data"], "timeout_seconds": 300}, "next_step": "call_assembler", "description": "Generate content for template placeholders", "output_field": "content_data"}, "call_wrapper": {"action": "call_agent", "config": {"agent_type": "multipage-wrapper", "target_role": "wrapper", "input_fields": ["final_html", "input_data"], "timeout_seconds": 60}, "next_step": "call_deployer", "description": "Create about and contact pages, package as files map", "output_field": "site_files"}, "spawn_writer": {"action": "spawn_agent", "config": {"role": "writer", "agent_type": "content-writer"}, "next_step": "spawn_assembler", "description": "Spawn content writer"}, "call_deployer": {"action": "call_agent", "config": {"agent_type": "site-deployer", "target_role": "deployer", "input_fields": ["site_files", "input_data"], "timeout_seconds": 180}, "next_step": "complete", "description": "Deploy to git repository", "output_field": "deployment_result"}, "spawn_wrapper": {"action": "spawn_agent", "config": {"role": "wrapper", "agent_type": "multipage-wrapper"}, "next_step": "spawn_deployer", "description": "Spawn multipage wrapper"}, "call_architect": {"action": "call_agent", "config": {"agent_type": "landing-page-architect", "target_role": "architect", "input_fields": ["build_plan", "brief_data", "input_data"], "timeout_seconds": 120}, "next_step": "call_writer", "description": "Assemble page template from components", "output_field": "template_data"}, "call_assembler": {"action": "call_agent", "config": {"agent_type": "html-assembler", "target_role": "assembler", "input_fields": ["content_data", "template_data", "brief_data", "input_data"], "timeout_seconds": 120}, "next_step": "call_wrapper", "description": "Assemble final HTML with CSS/JS", "output_field": "final_html"}, "spawn_deployer": {"action": "spawn_agent", "config": {"role": "deployer", "agent_type": "site-deployer"}, "next_step": "call_strategist", "description": "Spawn deployer"}, "call_strategist": {"action": "call_agent", "config": {"agent_type": "site-strategist", "target_role": "strategist", "input_fields": ["input_data", "brief_data"], "timeout_seconds": 120}, "next_step": "call_architect", "description": "Generate build plan from brief", "output_field": "build_plan"}, "spawn_architect": {"action": "spawn_agent", "config": {"role": "architect", "agent_type": "landing-page-architect"}, "next_step": "spawn_writer", "description": "Spawn landing page architect"}, "spawn_assembler": {"action": "spawn_agent", "config": {"role": "assembler", "agent_type": "html-assembler"}, "next_step": "spawn_wrapper", "description": "Spawn HTML assembler"}, "spawn_strategist": {"action": "spawn_agent", "config": {"role": "strategist", "agent_type": "site-strategist"}, "next_step": "spawn_architect", "description": "Spawn strategist"}}, "start_step": "spawn_strategist"}, "processing_mode": "orchestration", "timeout_seconds": 900}
is_active              | t
created_at             | 2025-12-04 08:42:32.269013+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["orchestration", "site-building", "landing-page"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {"sections": [{"name": "brand", "title": "Brand & Identity", "questions": [{"type": "text", "field": "brand_name", "label": "Brand/Company Name", "required": true}, {"type": "text", "field": "tagline", "label": "Tagline or Slogan", "required": false}, {"type": "select", "field": "tone", "label": "Brand Tone", "default": "professional", "options": ["professional", "friendly", "bold", "playful", "technical"]}]}, {"name": "value_proposition", "title": "Value Proposition", "questions": [{"type": "textarea", "field": "primary_benefit", "label": "What is the main benefit for visitors?", "required": true}, {"type": "textarea", "field": "unique_selling_points", "label": "What makes you different? (List 3-5 points)", "required": true}, {"type": "text", "field": "target_audience", "label": "Who is your ideal customer?", "required": true}]}, {"name": "conversion", "title": "Conversion Goals", "questions": [{"type": "text", "field": "primary_cta", "label": "Primary Call-to-Action (e.g., Sign Up, Buy Now)", "required": true}, {"type": "text", "field": "primary_cta_url", "label": "Primary CTA Link/Action", "required": false}, {"type": "text", "field": "secondary_cta", "label": "Secondary CTA (e.g., Learn More)", "required": false}]}, {"name": "social_proof", "title": "Trust & Social Proof", "questions": [{"type": "boolean", "field": "has_testimonials", "label": "Do you have customer testimonials?", "default": false}, {"type": "text", "field": "client_count", "label": "Number of customers/users (if applicable)", "required": false}, {"type": "text", "field": "notable_clients", "label": "Notable clients or partners", "required": false}]}]}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 33 ]---+--
id                     | a7091216-8361-492b-8a9b-60e5e9f4f5d8
type                   | multipage-wrapper
display_name           | Multi-Page Site Wrapper
description            | Wraps single-page site into multi-page structure (index, about, contact)
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return files map"}, "wrap_multipage": {"action": "wrap_multipage", "config": {"index_html_field": "final_html.final_html"}, "next_step": "complete", "description": "Create about and contact pages"}}, "start_step": "wrap_multipage"}, "processing_mode": "task", "timeout_seconds": 30}
is_active              | t
created_at             | 2025-12-02 18:41:21.558498+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["data-transformation", "html", "multipage"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"final_html": "object with html property"}, "required": ["final_html", "input_data"]}
output_contract        | {"format": {"type": "object", "description": "Map of filename to HTML content for all pages"}, "produces": "site_files"}
-[ RECORD 34 ]---+--
id                     | afbefbd4-2934-4131-9082-abac236eaa49
type                   | calculator
display_name           | Calculator Agent
description            | Performs mathematical calculations including addition, multiplication, and other operations
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return calculation result"}, "validate": {"action": "validate_input", "next_step": "calculate", "description": "Validate calculation request"}, "calculate": {"action": "calculate", "next_step": "complete", "description": "Execute the mathematical operation"}}, "start_step": "validate"}, "supported_operations": ["add", "subtract", "multiply", "divide", "modulo", "power"]}
is_active              | t
created_at             | 2025-09-09 14:45:22.264532+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["mathematical_operations", "arithmetic", "calculations"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 35 ]---+--
id                     | 21006167-913b-465b-90fa-fee45a88032b
type                   | copywriter
display_name           | Copywriter
description            | Creates compelling marketing and content copy
category               | data-driven
default_config         | {"model": "claude-3-sonnet", "ai_service": {"model": "claude-3-5-sonnet", "provider": "anthropic"}, "temperature": 0.7, "processing_mode": "task"}
is_active              | t
created_at             | 2025-08-10 16:00:36.153034+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | []
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"dlq": "system.agent.{type}.dlq", "errors": "system.agent.{type}.errors", "requests": "system.agent.{type}.requests", "responses": "system.agent.{type}.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | {"steps": {"execute": {"action": "execute_llm_prompt", "next_step": "complete", "description": "Execute task"}, "complete": {"action": "complete_workflow", "description": "Complete"}}, "start_step": "execute"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 36 ]---+--
id                     | 41677006-e536-4f78-a1b6-14c7f7916dff
type                   | site-deployer
display_name           | Site Deployer
description            | Commits assembled site to git repository. Works with intake orchestrator flow.
category               | deployment
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Deployment complete"}, "deploy_to_git": {"action": "git_commit", "config": {"repo_name": "sites", "files_field": "input_data.site_files.wrap_multipage.files", "domain_field": "input_data.domain", "content_field": "input_data.final_html.assemble_html.final_html", "commit_message": "Update site: {{.domain}}"}, "next_step": "complete", "description": "Commit pages to Git repository"}}, "start_step": "deploy_to_git"}, "processing_mode": "task", "timeout_seconds": 180}
is_active              | t
created_at             | 2025-11-28 11:17:59.582525+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["deployment", "git"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "250m", "memory": "256Mi"}, "requests": {"cpu": "50m", "memory": "128Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | integrator
status                 | active
domain_tags            | ["deployment", "git"]
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"site_files": "object (map of files)", "input_data.domain": "string", "input_data.repo_name": "string"}, "required": ["site_files", "input_data"]}
output_contract        | {"format": {"type": "object", "properties": {"url": "string", "status": "string", "commit_sha": "string"}, "description": "Deployment status and URLs"}, "produces": "deployment_result"}
-[ RECORD 37 ]---+--
id                     | 92b207b3-bf5e-4a4f-9fec-7b67d79f7678
type                   | content-creator-contact
display_name           | Contact Page Writer
description            | Specialized in writing welcoming and effective contact page content
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return contact page content"}, "generate_contact_content": {"action": "execute_llm_prompt", "config": {"input_fields": ["input_data"], "prompt_template": "{{.prompt}}\n\nCreate welcoming and clear content. Structure it with an introductory paragraph followed by contact details in a clean format. The contact details can be focused around how to contact the ai agents, or at least we can invent a customer contact (ai) agent. (There are not many humans involved in this venture)"}, "next_step": "complete", "description": "Generate contact page content"}}, "start_step": "generate_contact_content"}, "ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "max_tokens": 1500, "temperature": 0.7, "processing_mode": "task"}
is_active              | t
created_at             | 2025-10-30 20:10:19.155503+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["content", "contact", "communication", "customer-engagement"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.content-creator-contact", "process": "system.agent.content-creator-contact.process", "response": "system.responses.content-creator-contact"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 2
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | experimental
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | 
output_contract        | 
-[ RECORD 38 ]---+--
id                     | 5b3be002-256f-4f1f-88cf-446083a1738f
type                   | chief-strategist
display_name           | Chief Strategist V2
description            | Site planner that outputs pages with component types (v2 - unified architecture)
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return parsed plan"}, "parse_plan": {"action": "parse_json_field", "config": {"source_field": "build_plan_raw"}, "next_step": "complete", "description": "Parse JSON plan", "output_field": "plan_data"}, "generate_build_plan": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "output_type": "json", "input_fields": ["input_data"], "prompt_template": "You are a Site Planner designing the structure for {{.input_data.domain}}.\n\nOBJECTIVE: {{.input_data.objective}}\nMARKETING MODEL: {{.input_data.model}}\n\nSTEP 1: Determine the best site type for this objective.\n\nSite Type Guidelines:\n\nLANDING (1 page, 5-8 components)\n- Product launches, lead generation, focused campaigns\n- Single conversion goal, minimal navigation\n- Revenue: Direct sales, lead capture\n\nCORPORATE (4-6 pages)\n- Professional services, consulting, established businesses\n- Trust-building, multiple service areas\n- Revenue: Service contracts, B2B relationships\n\nPORTFOLIO (3-5 pages)\n- Creatives, agencies, freelancers\n- Case study focused, visual showcase\n- Revenue: Project work, client acquisition\n\nECOMMERCE (2-4 pages + product structure)\n- Product sales, shopping focused\n- Category browsing, cart functionality\n- Revenue: Direct product sales\n\nCONTENT (4-8 pages + article structure)\n- News sites, blogs, recipes, celebrity gossip, lifestyle\n- Content-driven traffic, regular publishing\n- SEO focused, high page count potential\n- Revenue: Advertising, affiliate links, sponsored content\n\nTOOLS (2-5 pages + tool interfaces)\n- Calculators (mortgage, tiles, BMI, etc.), converters, utilities\n- Feature/functionality driven, practical value\n- User retention through bookmarking\n- Revenue: Advertising, affiliate referrals, premium features\n\nSTEP 2: Plan each page with specific components.\n\nAvailable component types:\n- hero-centered, hero-split, hero-video\n- services-grid, services-list\n- features-cards, features-comparison\n- testimonials-carousel, testimonials-grid\n- team-grid, pricing-tiers, faq-accordion\n- cta-banner, cta-split\n- contact-form, contact-simple\n- about-story, about-values\n- footer-standard\n- blog-grid, blog-featured, article-layout\n- recipe-card, recipe-grid, recipe-detail\n- tool-calculator, tool-converter, tool-interface\n- ad-banner, ad-sidebar, affiliate-showcase\n- category-grid, content-feed, search-bar\n- social-share, comments-section, newsletter-signup\n\nSTEP 3: Create the sitemap with navigation structure.\n\nThe sitemap defines:\n- How each page appears in navigation\n- The URL for each page\n- Whether it appears in header nav, footer nav, or both\n\nURL Rules:\n- Home page: /index.html\n- Other pages: /{page-name}.html (e.g., /about.html, /services.html)\n- Use lowercase, hyphens for multi-word names\n\nOUTPUT FORMAT (valid JSON only):\n{\n  \"site_type\": \"landing|corporate|portfolio|ecommerce|content|tools\",\n  \"reasoning\": \"Why this structure fits the objective\",\n  \"theme_suggestion\": \"professional|bold|minimal|creative|editorial|functional\",\n  \"revenue_model\": \"direct_sales|services|advertising|affiliate|freemium\",\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Brand\",\n      \"purpose\": \"What this page achieves\",\n      \"components\": [\n        {\"type\": \"hero-centered\", \"priority\": \"high\"},\n        {\"type\": \"services-grid\", \"priority\": \"high\"}\n      ],\n      \"meta_description\": \"SEO description\"\n    }\n  ],\n  \"sitemap\": [\n    {\"label\": \"Home\", \"page\": \"index\", \"url\": \"/index.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"About\", \"page\": \"about\", \"url\": \"/about.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Services\", \"page\": \"services\", \"url\": \"/services.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Contact\", \"page\": \"contact\", \"url\": \"/contact.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Privacy Policy\", \"page\": \"privacy\", \"url\": \"/privacy.html\", \"in_header\": false, \"in_footer\": true}\n  ],\n  \"global\": {\n    \"brand_tone\": \"professional|friendly|bold|technical|editorial|practical\",\n    \"primary_cta\": {\"text\": \"Get Started\", \"url\": \"/contact.html\"}\n  }\n}\n\nIMPORTANT:\n- Every page in \"pages\" array MUST have a corresponding entry in \"sitemap\"\n- The \"page\" field in sitemap MUST match the \"name\" field in pages\n- Home page is always /index.html\n- Include privacy/terms pages in footer only (in_header: false)\n\nEXAMPLE - Corporate site:\n{\n  \"site_type\": \"corporate\",\n  \"pages\": [\n    {\"name\": \"index\", \"title\": \"Home | Acme Corp\", \"components\": [{\"type\": \"hero-centered\"}, {\"type\": \"services-grid\"}, {\"type\": \"testimonials-carousel\"}, {\"type\": \"cta-banner\"}]},\n    {\"name\": \"about\", \"title\": \"About Us | Acme Corp\", \"components\": [{\"type\": \"about-story\"}, {\"type\": \"team-grid\"}, {\"type\": \"about-values\"}]},\n    {\"name\": \"services\", \"title\": \"Our Services | Acme Corp\", \"components\": [{\"type\": \"hero-split\"}, {\"type\": \"services-list\"}, {\"type\": \"faq-accordion\"}]},\n    {\"name\": \"contact\", \"title\": \"Contact Us | Acme Corp\", \"components\": [{\"type\": \"contact-form\"}, {\"type\": \"cta-split\"}]}\n  ],\n  \"sitemap\": [\n    {\"label\": \"Home\", \"page\": \"index\", \"url\": \"/index.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"About Us\", \"page\": \"about\", \"url\": \"/about.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Services\", \"page\": \"services\", \"url\": \"/services.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Contact\", \"page\": \"contact\", \"url\": \"/contact.html\", \"in_header\": true, \"in_footer\": true},\n    {\"label\": \"Privacy Policy\", \"page\": \"privacy\", \"url\": \"/privacy.html\", \"in_header\": false, \"in_footer\": true}\n  ],\n  \"global\": {\n    \"brand_tone\": \"professional\",\n    \"primary_cta\": {\"text\": \"Get in Touch\", \"url\": \"/contact.html\"}\n  }\n}"}, "next_step": "parse_plan", "description": "Generate site plan with pages and components", "output_field": "build_plan_raw"}}, "start_step": "generate_build_plan"}, "processing_mode": "task", "timeout_seconds": 120}
is_active              | t
created_at             | 2025-11-19 13:49:41.175751+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["strategy"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | []
version                | 2
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | strategist
status                 | active
domain_tags            | ["website"]
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"input_data.domain": "string", "input_data.objective": "string"}, "required": ["input_data"]}
output_contract        | {"format": {"type": "object", "properties": {"pages": "array", "site_type": "string"}}, "produces": "plan_data"}
-[ RECORD 39 ]---+--
id                     | ae874fda-68b6-4830-8c5c-637fa3201dfc
type                   | content-creator
display_name           | Content Creator
description            | Advanced AI-powered content generation with memory and style adaptation
category               | data-driven
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "description": "Return website content"}, "create_content": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-haiku-4-5", "provider": "anthropic", "api_key_env_var": "ANTHROPIC_API_KEY"}, "output_type": "text", "input_fields": ["input_data", "domain_analysis", "site_architecture"], "prompt_template": "You are a professional copywriter. Create website content based on:\n\nDomain: {{.input_data.domain}}\nObjective: {{.input_data.objective}}\nPersuasion Model: {{.input_data.model}}\nDomain Analysis: {{.domain_analysis}}\nSite Architecture: {{.site_architecture}}\n\nCreate compelling content for each section. Use the specified persuasion model.\n\nReturn a JSON object with:\n- hero: headline, subheadline, cta_text\n- sections: array of section objects with title, content\n- meta: page_title, meta_description\n- footer: company info, links\n\nReturn ONLY valid JSON, no markdown or explanation."}, "next_step": "complete", "description": "Create website content", "output_field": "content_result"}}, "start_step": "create_content"}, "timeout_seconds": 300}
is_active              | t
created_at             | 2025-08-10 18:53:56.767957+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["content-generation", "writing", "copywriting", "seo", "memory-enabled", "style-adaptive"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "1000m", "memory": "2Gi"}, "requests": {"cpu": "200m", "memory": "512Mi"}}
topics                 | {"dlq": "system.agent.content-creator.dlq", "errors": "system.agent.content-creator.errors", "requests": "system.agent.content-creator.requests", "responses": "system.agent.content-creator.responses"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}
env_vars               | [{"name": "CONTENT_CREATOR_MODE", "value": "production"}, {"name": "ENABLE_METRICS", "value": "true"}, {"name": "MEMORY_ENABLED", "value": "true"}]
version                | 1
previous_version_id    | 
task_workflow          | {"steps": {"complete": {"action": "complete_workflow", "description": "Return generated content"}, "generate_content": {"action": "execute_llm_prompt", "config": {"prompt_template": "Generate engaging website content for: {{.business_type}}. Requirements: {{.requirements}}"}, "next_step": "complete", "description": "Generate website content"}}, "start_step": "generate_content"}
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": false}
agent_category         | executor
status                 | active
domain_tags            | ["content", "website"]
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"domain_analysis": "object", "site_architecture": "object"}, "required": ["input_data", "domain_analysis", "site_architecture"]}
output_contract        | {"format": {"type": "object", "description": "Content data for all site sections"}, "produces": "site_content"}
-[ RECORD 40 ]---+--
id                     | f7c8bee1-a845-4d5c-b136-761a844aba57
type                   | site-planner
display_name           | Site Planner
description            | Analyzes brief and plans site structure: pages, components, style collection, asset needs. Single LLM call to create comprehensive plan.
category               | specialist
default_config         | {"workflow": {"steps": {"complete": {"action": "complete_workflow", "config": {"output_field": "validated_plan"}, "description": "Return the validated site plan"}, "plan_site": {"action": "execute_llm_prompt", "config": {"ai_service": {"model": "claude-sonnet-4-5", "provider": "anthropic", "max_tokens": 4000, "api_key_env_var": "ANTHROPIC_API_KEY"}, "output_type": "json", "input_fields": ["input_data", "reviewed_brief", "available_components", "available_styles"], "prompt_template": "Plan a website for {{.input_data.domain}}.\n\n## Site Brief\n{{.reviewed_brief}}\n\n## Available Section Components\nThe following components are available in our component library. You MUST use ONLY these exact component names in the \"sections\" arrays:\n\n{{range .available_components}}\n- {{.name}} ({{.display_name}}): {{.function}} - {{.description}}\n{{end}}\n\n## Available Style Collections\n{{.available_styles}}\n\n## Task\nCreate a comprehensive site plan using ONLY the components listed above.\n\nReturn JSON in this format:\n```json\n{\n  \"pages\": [\n    {\n      \"name\": \"index\",\n      \"title\": \"Page Title | Site Name\",\n      \"nav_label\": \"Home\",\n      \"nav_order\": 1,\n      \"in_header\": true,\n      \"in_footer\": true,\n      \"sections\": [\"hero\", \"features\", \"testimonials\", \"call_to_action\"]\n    }\n  ],\n  \"style_collection\": \"style-name\",\n  \"needs_logo\": true,\n  \"needs_images\": true,\n  \"image_prompts\": {\n    \"logo\": \"Description for logo generation\",\n    \"hero_home\": \"Description for home hero image\"\n  }\n}\n```\n\nSTRICT RULES:\n1. ONLY use component names from the \"Available Section Components\" list above\n2. DO NOT invent new component names - if unsure, use \"hero\" for hero sections, \"features\" for feature lists, \"call_to_action\" for CTAs\n3. Use these standard mappings:\n   - For any hero/banner at page top: use \"hero\" or page-specific variants like \"contact-hero\", \"services-hero\", \"about-hero\"\n   - For feature lists: use \"features\"\n   - For service listings: use \"services-grid\"\n   - For testimonials/quotes: use \"testimonials\" or \"social_proof\"\n   - For calls to action: use \"call_to_action\"\n   - For contact forms: use \"contact-form\"\n   - For contact details: use \"contact-info\"\n   - For team sections: use \"leadership-team\"\n   - For case studies: use \"case-studies-list\"\n   - For about content: use \"about-content\"\n   - For differentiators/why-us: use \"differentiators-section\"\n\n4. Choose style_collection based on industry and tone from the brief\n5. Keep header navigation to 5-8 items maximum\n6. Always include: index (home) and contact pages\n\nIMAGE GENERATION (REQUIRED - DO NOT SKIP):\nYou MUST include needs_logo, needs_images, and image_prompts in your response.\n\n- Set needs_logo: true (always)\n- Set needs_images: true (always)\n- Provide image_prompts object with BOTH of these keys:\n  - \"logo\": A detailed 2-3 sentence prompt for logo generation. Describe the style (modern, classic, minimal), colors that match the brand, and any relevant imagery for the industry.\n  - \"hero_home\": A detailed 2-3 sentence prompt for the homepage hero background. Describe the mood (professional, energetic, calm), imagery type (abstract, photographic, geometric), and colors/atmosphere that fit the brand.\n\nExample image_prompts:\n{\n  \"logo\": \"A modern, minimal logo for a tech consulting company. Use clean geometric shapes with a navy blue and teal color palette. The design should convey innovation and trustworthiness.\",\n  \"hero_home\": \"A professional, abstract background with flowing gradients in deep navy and teal. Include subtle geometric patterns that suggest technology and connectivity. The mood should be confident and forward-thinking.\"\n}"}, "next_step": "validate_plan", "description": "LLM creates comprehensive site plan", "output_field": "llm_plan"}, "validate_plan": {"action": "validate_site_plan", "config": {"max_pages": 20, "plan_field": "llm_plan.result", "ensure_pages": ["index", "contact"], "default_style": "professional-dark", "validate_components": true}, "next_step": "complete", "description": "Validate and normalize the site plan", "output_field": "validated_plan"}, "load_style_collections": {"action": "query_database", "config": {"query": "SELECT name, display_name, category, description FROM style_collections WHERE is_active = true ORDER BY name", "output_format": "array"}, "next_step": "plan_site", "description": "Load available style collections", "output_field": "available_styles"}, "load_available_components": {"action": "query_database", "config": {"query": "SELECT name, display_name, \"function\", category, description FROM content_components WHERE component_level IN ('section', 'element') AND is_active = true ORDER BY category, name", "output_format": "array"}, "next_step": "load_style_collections", "description": "Load available section components from database", "output_field": "available_components"}}, "start_step": "load_available_components"}, "processing_mode": "task", "timeout_seconds": 120}
is_active              | t
created_at             | 2025-12-22 17:47:04.733638+00
updated_at             | 2026-01-22 09:47:43.43281+00
deleted_at             | 
capabilities           | ["planning", "site-architecture", "component-selection"]
image_repository       | docker.io/aqls/agent-chassis
image_tag              | v1.0.696
command                | 
resources              | {"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}
topics                 | {"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}
health_config          | {"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 15}
env_vars               | []
version                | 1
previous_version_id    | 
task_workflow          | 
orchestrator_workflow  | 
orchestration_workflow | 
delegation_preferences | {"fallback_to_self": true, "prefer_delegation": true}
agent_category         | 
status                 | active
domain_tags            | []
briefing_questionnaire | {}
usage_count            | 0
is_snapshot            | f
input_contract         | {"expects": {"input_data": {"domain": "string - the domain name", "objective": "string - what the site should achieve"}, "reviewed_brief": "object - questionnaire responses (structure varies by site type)"}, "required": ["input_data.domain", "reviewed_brief"]}
output_contract        | {"produces": {"pages": "array of {name, title, nav_label, nav_order, sections[], in_header, in_footer}", "needs_logo": "boolean - whether to generate a logo", "needs_images": "boolean - whether to generate hero/background images", "image_prompts": "object with prompts for each needed image", "style_collection": "string - name of style collection to use"}}
-[ RECORD 41 ]---+--
id                     | d77eecc2-b73c-4ce2-8a86-ff20f0c93063
type                   | brand-designer
display_name           | Brand Designer Agent
description            | Analyzes domain, industry, and objectives to select or generate custom CSS themes and brand guidelines
                                                                                                                                                                                                                                                                                                                        category               | data-driven
                                                                                                                                                                                                                                                                                                                   output_contract        | 
