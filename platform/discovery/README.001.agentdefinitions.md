-- Welcome Message Generator
INSERT INTO agent_group_definitions (
id,
name,
group_type,
version,
agent_configs,
orchestration_workflow,
usage_count
) VALUES (
gen_random_uuid(),
'Welcome Message Generator',
'welcome-message-generator',
1,
'[
{
"role": "writer",
"agent_type": "content-creator"
}
]'::jsonb,
'{
"start_step": "spawn_writer",
"steps": {
"spawn_writer": {
"action": "spawn_agent",
"config": {
"agent_type": "content-creator",
"role": "writer"
},
"next_step": "create_welcome"
},
"create_welcome": {
"action": "call_agent",
"config": {
"target_role": "writer"
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow"
}
}
}'::jsonb,
0
);

-- Simple Website Builder
INSERT INTO agent_group_definitions (
id,
name,
group_type,
version,
agent_configs,
orchestration_workflow,
usage_count
) VALUES (
gen_random_uuid(),
'Simple Website Builder',
'simple-website-builder',
1,
'[
{
"role": "content_writer",
"agent_type": "content-creator"
},
{
"role": "html_coder",
"agent_type": "html-developer"
}
]'::jsonb,
'{
"start_step": "spawn_content_writer",
"steps": {
"spawn_content_writer": {
"action": "spawn_agent",
"config": {
"agent_type": "content-creator",
"role": "content_writer"
},
"next_step": "spawn_html_coder"
},
"spawn_html_coder": {
"action": "spawn_agent",
"config": {
"agent_type": "html-developer",
"role": "html_coder"
},
"next_step": "create_content"
},
"create_content": {
"action": "call_agent",
"config": {
"target_role": "content_writer"
},
"next_step": "develop_html"
},
"develop_html": {
"action": "call_agent",
"config": {
"target_role": "html_coder"
},
"next_step": "store_website"
},
"store_website": {
"action": "upload_to_s3",
"config": {
"make_public": true
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow"
}
}
}'::jsonb,
0
);