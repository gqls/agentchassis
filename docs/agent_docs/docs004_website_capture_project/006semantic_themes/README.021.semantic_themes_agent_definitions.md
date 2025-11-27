-- ============================================================================
-- REFACTORED CONTENT CREATOR - Outputs structured JSON, not full HTML
-- ============================================================================

-- Update the content-creator agent definition
UPDATE agent_definitions
SET
updated_at = now(),
default_config = '{
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
"input_fields": ["template_data", "build_plan_data", "input_data"],
"output_field": "content_json",
"prompt_template": "You are a professional website content creator. Your job is to create compelling, industry-specific CONTENT (not HTML structure).\n\nWebsite Details:\n- Domain: {{.domain}}\n- Objective: {{.objective}}\n- Model: {{.model}}\n\nBuild Strategy (from strategist):\n{{.build_plan_data.generate_build_plan.result}}\n\nContent Requirements - these are the placeholders you need to fill:\n{{.template_data.assemble_template.content_requirements}}\n\nYour Task:\nCreate a JSON object with content for each placeholder. Group by component.\n\nGuidelines:\n- Write compelling, conversion-focused copy\n- Match the domain and industry tone\n- For testimonials, use optimistic placeholder attributions like \"[Future You]\", \"[Soon-to-be Satisfied Customer]\" - NOT fake names\n- Use action-oriented language for CTAs\n- Keep brand consistency throughout\n- Stats/numbers should be realistic placeholders like \"500+\" or \"10,000+\"\n\nReturn ONLY valid JSON in this exact structure:\n{\n  \"meta\": {\n    \"title\": \"Page title for browser tab\",\n    \"description\": \"SEO meta description (150-160 chars)\"\n  },\n  \"theme\": \"recommended theme name from: default, calm-minimal, bold-conversion, warm-friendly, dark-modern, premium-elegant\",\n  \"theme_tags\": [\"semantic\", \"tags\", \"for\", \"theme\", \"matching\"],\n  \"sections\": {\n    \"component_header_0\": {\n      \"brand_name\": \"Your Brand Name\",\n      \"cta_text\": \"CTA Button Text\"\n    },\n    \"component_hero_1\": {\n      \"headline\": \"Main headline\",\n      \"subheadline\": \"Supporting text\",\n      \"primary_cta\": \"Primary button\",\n      \"secondary_cta\": \"Secondary button\"\n    }\n  }\n}\n\nFill ALL placeholders from the content requirements. Return ONLY the JSON object, no markdown or explanation."
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return the content JSON",
"output_field": ["content_json"]
}
}
},
"processing_mode": "task",
"timeout_seconds": 300
}'::jsonb
WHERE type = 'content-creator';

-- ============================================================================
-- NEW: HTML ASSEMBLER AGENT
-- Takes content JSON + template and produces final HTML with CSS/JS
-- ============================================================================

INSERT INTO agent_definitions (
id, type, display_name, description, category,
default_config, is_active, capabilities,
image_repository, image_tag, resources, topics, health_config
)
VALUES (
gen_random_uuid(),
'html-assembler',
'HTML Assembler Agent',
'Assembles final HTML from content JSON, templates, CSS themes, and JS snippets',
'data-driven',
'{
"workflow": {
"start_step": "assemble_html",
"steps": {
"assemble_html": {
"action": "assemble_full_page",
"config": {
"input_fields": ["content_data", "template_data", "input_data"],
"include_css_snippets": true,
"include_js_snippets": true,
"minify": false
},
"next_step": "complete",
"description": "Assemble complete HTML page with CSS and JS"
},
"complete": {
"action": "complete_workflow",
"description": "Return the assembled HTML"
}
}
},
"processing_mode": "task",
"timeout_seconds": 120
}'::jsonb,
true,
ARRAY['html', 'assembly', 'css', 'javascript'],
'docker.io/aqls/agent-chassis',
'v1.0.476',
'{"limits": {"cpu": "500m", "memory": "1Gi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
'{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
'{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb
)
ON CONFLICT (type) DO UPDATE SET
default_config = EXCLUDED.default_config,
updated_at = now();

-- ============================================================================
-- BRIEFING AGENT
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
'Analyzes user input and generates structured brief for strategy generation',
'data-driven',
'{
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
"prompt_template": "Analyze this website request and create a comprehensive brief for the strategy and content teams.\n\nInput:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Model (if specified): {{.input_data.model}}\n\nAnalyze the domain name and objective to infer:\n1. The likely industry/niche\n2. Target audience demographics and psychographics\n3. Appropriate tone and style\n4. Key messages that should be communicated\n5. Unique selling propositions to emphasize\n6. Recommended page sections\n7. Best matching theme style\n\nReturn a JSON brief:\n{\n  \"analysis\": {\n    \"industry\": \"detected industry/niche\",\n    \"industry_confidence\": 0.0-1.0,\n    \"domain_interpretation\": \"what the domain name suggests\"\n  },\n  \"audience\": {\n    \"primary\": \"primary target audience description\",\n    \"secondary\": \"secondary audience if applicable\",\n    \"demographics\": [\"age range\", \"profession\", \"etc\"],\n    \"psychographics\": [\"values\", \"motivations\", \"pain points\"]\n  },\n  \"brand\": {\n    \"tone\": \"professional|casual|technical|friendly|authoritative|playful\",\n    \"personality\": [\"trait1\", \"trait2\", \"trait3\"],\n    \"voice_examples\": [\"example phrase in brand voice\"]\n  },\n  \"messaging\": {\n    \"value_proposition\": \"core value proposition in one sentence\",\n    \"key_messages\": [\"message1\", \"message2\", \"message3\"],\n    \"usps\": [\"unique selling point 1\", \"usp 2\", \"usp 3\"],\n    \"proof_points\": [\"credibility element 1\", \"element 2\"]\n  },\n  \"structure\": {\n    \"recommended_sections\": [\"header\", \"hero\", \"social_proof\", \"features\", \"pricing\", \"faq\", \"cta\", \"footer\"],\n    \"priority_sections\": [\"most important sections for this objective\"],\n    \"optional_sections\": [\"sections that could be added\"]\n  },\n  \"theme\": {\n    \"recommended\": \"theme name\",\n    \"semantic_tags\": [\"tag1\", \"tag2\", \"tag3\"],\n    \"color_mood\": \"description of appropriate color feeling\",\n    \"style_notes\": \"any specific style recommendations\"\n  },\n  \"content_guidelines\": {\n    \"headline_style\": \"guidance for headlines\",\n    \"cta_style\": \"guidance for calls to action\",\n    \"avoid\": [\"things to avoid in copy\"],\n    \"emphasize\": [\"things to emphasize\"]\n  }\n}\n\nReturn ONLY valid JSON, no markdown or explanation."
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Return the structured brief"
}
}
},
"processing_mode": "task",
"timeout_seconds": 60
}'::jsonb,
true,
ARRAY['analysis', 'briefing', 'strategy', 'llm'],
'docker.io/aqls/agent-chassis',
'v1.0.476',
'{"limits": {"cpu": "500m", "memory": "512Mi"}, "requests": {"cpu": "100m", "memory": "256Mi"}}'::jsonb,
'{"error": "system.errors.{type}", "process": "system.agent.{type}.process", "response": "system.responses.{type}"}'::jsonb,
'{"port": 8080, "liveness_path": "/health", "readiness_path": "/ready", "initial_delay_seconds": 30}'::jsonb
)
ON CONFLICT (type) DO UPDATE SET
default_config = EXCLUDED.default_config,
updated_at = now();

-- ============================================================================
-- UPDATED MVP-SITE-BUILDER WORKFLOW
-- Now includes: Briefing → Strategist → Architect → Content → Assembler → Deployer
-- ============================================================================

UPDATE agent_group_definitions
SET
updated_at = now(),
agent_configs = '[
{"role": "briefer", "agent_type": "briefing-agent"},
{"role": "chief_strategist", "agent_type": "chief-strategist"},
{"role": "site_component_architect", "agent_type": "site-component-architect"},
{"role": "content_creator", "agent_type": "content-creator"},
{"role": "html_assembler", "agent_type": "html-assembler"},
{"role": "deployer", "agent_type": "deployer-agent"}
]'::jsonb,
orchestration_workflow = '{
"start_step": "spawn_briefer",
"steps": {
"spawn_briefer": {
"action": "spawn_agent",
"config": {
"role": "briefer",
"agent_type": "briefing-agent"
},
"next_step": "spawn_strategist",
"description": "Spawn Briefing Agent"
},
"spawn_strategist": {
"action": "spawn_agent",
"config": {
"role": "chief_strategist",
"agent_type": "chief-strategist"
},
"next_step": "spawn_architect",
"description": "Spawn Chief Strategist"
},
"spawn_architect": {
"action": "spawn_agent",
"config": {
"role": "site_component_architect",
"agent_type": "site-component-architect"
},
"next_step": "spawn_content_creator",
"description": "Spawn Site Component Architect"
},
"spawn_content_creator": {
"action": "spawn_agent",
"config": {
"role": "content_creator",
"agent_type": "content-creator"
},
"next_step": "spawn_assembler",
"description": "Spawn Content Creator"
},
"spawn_assembler": {
"action": "spawn_agent",
"config": {
"role": "html_assembler",
"agent_type": "html-assembler"
},
"next_step": "spawn_deployer",
"description": "Spawn HTML Assembler"
},
"spawn_deployer": {
"action": "spawn_agent",
"config": {
"role": "deployer",
"agent_type": "deployer-agent"
},
"next_step": "call_briefer",
"description": "Spawn Deployer"
},
"call_briefer": {
"action": "call_agent",
"description": "Get structured brief from user input",
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
"action": "call_agent",
"description": "Build the empty template from the Build Plan",
"config": {
"agent_type": "site-component-architect",
"target_role": "site_component_architect",
"input_fields": ["build_plan_data", "brief_data", "input_data"],
"timeout_seconds": 120
},
"output_field": "template_data",
"next_step": "call_content_creator"
},
"call_content_creator": {
"action": "call_agent",
"description": "Generate content JSON for the template",
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
description = '[MVP v2] 6-step workflow: Briefing → Strategy → Architecture → Content → Assembly → Deploy'
WHERE group_type = 'mvp-site-builder';

-- ============================================================================
-- UPDATE DEPLOYER to use new field path
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
"domain_field": "domain",
"input_fields": ["input_data"],
"content_field": "input_data.final_html_data.assemble_html.final_html",
"commit_message": "Update site: {{.domain}}",
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
WHERE type = 'deployer-agent';


========================
========================

Chief strategist update · SQL
-- ============================================================================
-- UPDATED CHIEF STRATEGIST - Now uses briefing data for enhanced strategy
-- ============================================================================

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
"max_tokens": 3000
},
"input_fields": ["input_data", "brief_data"],
"output_field": "build_plan_json",
"prompt_template": "You are a website strategist creating a Build Plan based on behavioral psychology and conversion optimization.\n\nWebsite Request:\n- Domain: {{.input_data.domain}}\n- Objective: {{.input_data.objective}}\n- Behavioral Model: {{.input_data.model}}\n\n{{if .brief_data}}Detailed Brief (from analysis):\n{{.brief_data.structured_brief.result}}\n{{end}}\n\nYour Task:\nCreate a strategic Build Plan that maps behavioral psychology to website sections.\n\nBehavioral Models Available:\n- PAS (Problem-Agitate-Solution): Best for pain-point focused products\n- AIDA (Attention-Interest-Desire-Action): Best for general conversion\n- FAB (Features-Advantages-Benefits): Best for feature-rich products\n- 4Ps (Promise-Picture-Proof-Push): Best for aspirational products\n\nFor each model, define the appropriate sections:\n\nPAS sections: [\"problem_statement\", \"agitation\", \"solution_provider\", \"social_proof\", \"cta\"]\nAIDA sections: [\"attention_hero\", \"interest_features\", \"desire_benefits\", \"action_cta\"]\nFAB sections: [\"features_showcase\", \"advantages_comparison\", \"benefits_outcome\", \"cta\"]\n4Ps sections: [\"promise_hero\", \"picture_vision\", \"proof_testimonials\", \"push_cta\"]\n\n{{if .brief_data}}Use the briefing analysis to enhance your section selection:\n- Recommended theme: Use the theme.recommended from the brief\n- Messaging: Incorporate key_messages and usps into section guidance\n- Audience: Consider the target audience when defining section priorities\n{{end}}\n\nReturn ONLY valid JSON with this structure:\n{\n  \"model\": \"PAS|AIDA|FAB|4Ps\",\n  \"sections\": [\"section1\", \"section2\", ...],\n  \"section_guidance\": {\n    \"section_name\": {\n      \"purpose\": \"what this section should achieve\",\n      \"key_message\": \"primary message for this section\",\n      \"tone\": \"section-specific tone guidance\"\n    }\n  },\n  \"theme_recommendation\": \"recommended theme name\",\n  \"theme_tags\": [\"semantic\", \"tags\"],\n  \"conversion_priority\": [\"most important sections for conversion\"]\n}\n\nDO NOT include any text outside the JSON object."
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
WHERE type = 'chief-strategist';


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
-- "build_plan_field": "build_plan_data",
"input_fields": ["build_plan_data"]
},
"next_step": "complete",
"description": "Assemble HTML template from component library"
},
"complete": {
"action": "complete_workflow",
"description": "Return the assembled template"
}
}
},
"processing_mode": "task",
"timeout_seconds": 180,
}'::jsonb
WHERE type = 'site-component-architect';