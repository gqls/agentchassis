-- Website Builder Orchestrator Agent Definition
-- This is the master orchestrator that coordinates all website building activities

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
    orchestration_workflow,
    delegation_preferences
) VALUES (
             gen_random_uuid(),
             'website-builder-orchestrator',
             'Website Builder Orchestrator',
             'Master orchestrator for building websites from domain analysis to final output',
             'orchestration',
             '{
                 "workflow": {
                     "start_step": "analyze_input",
                     "steps": {
                         "analyze_input": {
                             "action": "analyze_input_type",
                             "description": "Determine if input is domain, URL, or design brief",
                             "config": {
                                 "analysis_type": "input_classification"
                             },
                             "next_step": "spawn_capture_agent"
                         },
                         "spawn_capture_agent": {
                             "action": "spawn_agent",
                             "description": "Spawn agent for website capture",
                             "config": {
                                 "agent_type": "website-capture",
                                 "role": "capture_specialist"
                             },
                             "next_step": "capture_website_data"
                         },
                         "capture_website_data": {
                             "action": "call_agent",
                             "description": "Capture screenshots and HTML/CSS",
                             "config": {
                                 "agent_type": "website-capture",
                                 "target_role": "capture_specialist",
                                 "prompt": "Capture website data for {{.input_data.target_url}}",
                                 "timeout_seconds": 120
                             },
                             "next_step": "spawn_vision_agent"
                         },
                         "spawn_vision_agent": {
                             "action": "spawn_agent",
                             "description": "Spawn visual analysis agent",
                             "config": {
                                 "agent_type": "website-vision",
                                 "role": "vision_analyst"
                             },
                             "next_step": "analyze_visuals"
                         },
                         "analyze_visuals": {
                             "action": "call_agent",
                             "description": "Analyze visual components and layout",
                             "config": {
                                 "agent_type": "website-vision",
                                 "target_role": "vision_analyst",
                                 "input_data": {
                                     "screenshot_path": "{{.capture_website_data.screenshot_path}}",
                                     "mobile_screenshot_path": "{{.capture_website_data.mobile_screenshot_path}}"
                                 }
                             },
                             "next_step": "spawn_code_agent"
                         },
                         "spawn_code_agent": {
                             "action": "spawn_agent",
                             "description": "Spawn code analysis agent",
                             "config": {
                                 "agent_type": "website-code-analyzer",
                                 "role": "code_analyst"
                             },
                             "next_step": "analyze_code"
                         },
                         "analyze_code": {
                             "action": "call_agent",
                             "description": "Clean and analyze HTML/CSS structure",
                             "config": {
                                 "agent_type": "website-code-analyzer",
                                 "target_role": "code_analyst",
                                 "input_data": {
                                     "html_content": "{{.capture_website_data.html_content}}",
                                     "css_content": "{{.capture_website_data.css_content}}"
                                 }
                             },
                             "next_step": "spawn_synthesis_agent"
                         },
                         "spawn_synthesis_agent": {
                             "action": "spawn_agent",
                             "description": "Spawn synthesis agent",
                             "config": {
                                 "agent_type": "website-synthesis",
                                 "role": "synthesizer"
                             },
                             "next_step": "synthesize_design"
                         },
                         "synthesize_design": {
                             "action": "call_agent",
                             "description": "Correlate visual and code analysis",
                             "config": {
                                 "agent_type": "website-synthesis",
                                 "target_role": "synthesizer",
                                 "input_data": {
                                     "visual_map": "{{.analyze_visuals.visual_map}}",
                                     "cleaned_structure": "{{.analyze_code.cleaned_structure}}",
                                     "color_palette": "{{.analyze_visuals.color_palette}}"
                                 }
                             },
                             "next_step": "spawn_content_strategist"
                         },
                         "spawn_content_strategist": {
                             "action": "spawn_agent",
                             "description": "Spawn content strategy agent",
                             "config": {
                                 "agent_type": "content-strategist",
                                 "role": "content_strategist"
                             },
                             "next_step": "plan_content"
                         },
                         "plan_content": {
                             "action": "call_agent",
                             "description": "Plan content structure for new site",
                             "config": {
                                 "agent_type": "content-strategist",
                                 "target_role": "content_strategist",
                                 "input_data": {
                                     "business_type": "{{.input_data.business_type}}",
                                     "business_name": "{{.input_data.business_name}}",
                                     "template_structure": "{{.synthesize_design.template}}"
                                 }
                             },
                             "next_step": "generate_sections"
                         },
                         "generate_sections": {
                             "action": "parallel_section_generation",
                             "description": "Generate all website sections in parallel",
                             "config": {
                                 "sections": ["hero", "features", "testimonials", "about", "contact", "cta"],
                                 "content_plan": "{{.plan_content.content_plan}}"
                             },
                             "next_step": "aggregate_website"
                         },
                         "aggregate_website": {
                             "action": "aggregate_webpage",
                             "description": "Combine all sections into final website",
                             "config": {
                                 "template": "{{.synthesize_design.template}}",
                                 "styles": "{{.synthesize_design.styles}}",
                                 "sections": "{{.generate_sections}}",
                                 "output_format": "complete_website"
                             },
                             "next_step": "store_in_library"
                         },
                         "store_in_library": {
                             "action": "store_component",
                             "description": "Store website and components in library",
                             "config": {
                                 "storage_type": "postgres_vector",
                                 "include_components": true,
                                 "generate_embeddings": true
                             },
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Return complete website package"
                         }
                     }
                 },
                 "processing_mode": "orchestration",
                 "ai_service": {
                     "provider": "anthropic",
                     "model": "claude-3-5-sonnet-20241022",
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "max_tokens": 4000,
                 "temperature": 0.7
             }'::jsonb,
             true,
             ARRAY['orchestration', 'website-building', 'coordination'],
             'docker.io/aqls/agent-chassis',
             'v1.0.407',
             NULL,
             '{"prefer_delegation": true, "fallback_to_self": false}'::jsonb
         );