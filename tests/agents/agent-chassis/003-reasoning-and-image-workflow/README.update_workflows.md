-- Fix the Creative Orchestrator (remove the circular dependency)
UPDATE client_demo_client.agent_instances
SET config = jsonb_set(
config,
'{workflow,steps,generate_description}',
jsonb_build_object(
'topic', 'system.agent.reasoning.process',
'action', 'reason_about_image',
'next_step', 'create_image',
'description', 'Use reasoning agent to create detailed image description'
)
)
WHERE id = '00000000-0000-0000-0000-000000000010';

-- Fix the Multi-Agent Orchestrator (change subtasks to sub_tasks)
UPDATE client_demo_client.agent_instances
SET config = jsonb_set(
config,
'{workflow,steps,fan_out_to_agents}',
jsonb_build_object(
'action', 'fan_out',
'sub_tasks', ARRAY[
jsonb_build_object('topic', 'system.agent.web-search.process', 'step_name', 'web_search'),
jsonb_build_object('topic', 'system.agent.content-creator.process', 'step_name', 'content_creation'),
jsonb_build_object('topic', 'system.agent.copywriter.process', 'step_name', 'copywriting')
]::jsonb[],
'next_step', 'complete',
'description', 'Call multiple agents in parallel'
)
)
WHERE id = '00000000-0000-0000-0000-000000000040';

-- Verify the fixes
SELECT id, name, config->'workflow'->'steps' as steps
FROM client_demo_client.agent_instances
WHERE id IN ('00000000-0000-0000-0000-000000000010', '00000000-0000-0000-0000-000000000040');