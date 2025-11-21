INSERT INTO agent_definitions (type, display_name, description, category, default_config)
VALUES (
'brand-designer',
'Brand Designer Agent',
'Analyzes domain, industry, and objectives to select or generate custom CSS themes and brand guidelines',
'data-driven',
'{
"workflow": {
"start_step": "analyze_brand",
"steps": {
"analyze_brand": {
"action": "execute_llm_prompt",
"config": {
"ai_service": {
"provider": "anthropic",
"model": "claude-haiku-4-5-20251001",
"api_key_env_var": "ANTHROPIC_API_KEY"
},
"prompt_template": "Analyze {{.domain}} and {{.objective}}. Choose the most appropriate theme from: boxing, bakery, tech, professional-dark, default. Consider industry, target audience, and brand positioning. Return JSON: {\"theme\": \"...\", \"reasoning\": \"...\"}",
"output_field": "brand_analysis"
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow"
}
}
},
"processing_mode": "task",
"timeout_seconds": 90
}'
);


UPDATE agent_group_definitions
SET orchestration_workflow = '{
"start_step": "spawn_strategist",
"steps": {
"spawn_strategist": {...},
"spawn_brand_designer": {
"action": "spawn_agent",
"config": {
"role": "brand_designer",
"agent_type": "brand-designer"
},
"next_step": "spawn_architect"
},
"call_brand_designer": {
"action": "call_agent",
"config": {
"agent_type": "brand-designer",
"target_role": "brand_designer",
"input_fields": ["input_data"],
"timeout_seconds": 60
},
"next_step": "call_architect",
"output_field": "brand_theme"
},
"call_architect": {
"action": "call_agent",
"config": {...},
"input_fields": ["generate_build_plan", "brand_theme", "input_data"]
},
...
}
}'
WHERE group_type = 'mvp-site-builder';