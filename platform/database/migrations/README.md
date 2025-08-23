-- additional

kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db << 'EOF'
-- Update all agent definitions to include storage environment variables
UPDATE agent_definitions
SET env_vars = '[
{"name": "AWS_ACCESS_KEY_ID", "valueFrom": {"secretKeyRef": {"name": "personae-storage-secrets", "key": "AWS_ACCESS_KEY_ID"}}},
{"name": "AWS_SECRET_ACCESS_KEY", "valueFrom": {"secretKeyRef": {"name": "personae-storage-secrets", "key": "AWS_SECRET_ACCESS_KEY"}}},
{"name": "S3_ENDPOINT", "value": "https://s3.us-east-005.backblazeb2.com"},
{"name": "S3_REGION", "value": "us-west-004"},
{"name": "IMAGE_BUCKET", "value": "personae-prod-uk001-images"},
{"name": "ASSETS_BUCKET", "value": "personae-prod-uk001-site-assets"},
{"name": "S3_USE_PATH_STYLE", "value": "false"}
]'::jsonb
WHERE type IN ('site-publisher', 'html-developer', 'visual-designer');

-- Also update the default_config to include storage settings
UPDATE agent_definitions
SET default_config = default_config || '{
"storage": {
"provider": "s3",
"endpoint": "https://s3.us-east-005.backblazeb2.com",
"bucket": "personae-prod-uk001-site-assets",
"access_key_env_var": "AWS_ACCESS_KEY_ID",
"secret_key_env_var": "AWS_SECRET_ACCESS_KEY"
}
}'::jsonb
WHERE type IN ('site-publisher', 'html-developer', 'visual-designer');

SELECT type, env_vars FROM agent_definitions WHERE type IN ('site-publisher', 'html-developer');
EOF


---
 mark s3 upload as local action

kubectl -n ai-persona-system exec -it postgres-templates-0 -- psql -U templates_user -d templates_db -c "
UPDATE agent_definitions
SET default_config = jsonb_set(
default_config,
'{workflow}',
'{
\"start_step\": \"collect_assets\",
\"steps\": {
\"collect_assets\": {
\"action\": \"validate_input\",
\"next_step\": \"organize_files\",
\"description\": \"Validate and collect website assets\"
},
\"organize_files\": {
\"action\": \"transform_data\",
\"next_step\": \"upload_to_bucket\",
\"description\": \"Organize files for upload\"
},
\"upload_to_bucket\": {
\"action\": \"s3_upload\",
\"config\": {
\"is_local_action\": true
},
\"next_step\": \"complete\",
\"description\": \"Upload website to S3 bucket\"
},
\"complete\": {
\"action\": \"complete_workflow\",
\"description\": \"Complete the publishing process\"
}
}
}'::jsonb
)
WHERE type = 'site-publisher';"

--

$ kubectl -n ai-persona-system exec -it postgres-templates-0 -- psql -U templates_user -d templates_db -c "
CREATE TABLE IF NOT EXISTS agent_definitions (
id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
type VARCHAR(255) UNIQUE NOT NULL,
display_name VARCHAR(255) NOT NULL,
description TEXT,
category VARCHAR(100),
image_repository VARCHAR(255) DEFAULT 'docker.io/aqls/agent-chassis',
image_tag VARCHAR(100) DEFAULT 'latest',
command TEXT[] DEFAULT ARRAY['./agent-chassis', '-config', 'configs/agent-chassis.yaml'],
resources JSONB DEFAULT '{\"requests\": {\"cpu\": \"100m\", \"memory\": \"256Mi\"}, \"limits\": {\"cpu\": \"500m\", \"memory\": \"1Gi\"}}'::jsonb,
default_config JSONB DEFAULT '{}'::jsonb,
capabilities JSONB DEFAULT '[]'::jsonb,
topics JSONB DEFAULT '{}'::jsonb,
health_config JSONB DEFAULT '{\"liveness_path\": \"/health\", \"readiness_path\": \"/ready\", \"port\": 8080, \"initial_delay_seconds\": 30}'::jsonb,
env_vars JSONB DEFAULT '[]'::jsonb,
is_active BOOLEAN DEFAULT true,
deleted_at TIMESTAMP,
created_at TIMESTAMP DEFAULT NOW(),
updated_at TIMESTAMP DEFAULT NOW()
);"
CREATE TABLE


--

kubectl -n ai-persona-system exec -it postgres-templates-0 -- psql -U templates_user -d templates_db -c "
-- Website Builder
INSERT INTO agent_definitions (type, display_name, description, category, capabilities, topics, default_config) VALUES
('website-builder', 'Website Builder', 'Orchestrates website creation', 'orchestrator',
'[\"website_creation\", \"orchestration\", \"multi_agent_coordination\"]'::jsonb,
'{\"process\": \"system.agent.website-builder.process\", \"response\": \"system.responses.website-builder\", \"error\": \"system.errors.website-builder\", \"dlq\": \"dlq.website-builder\"}'::jsonb,
'{\"workflow\": {\"start_step\": \"analyze_domain\", \"steps\": {\"analyze_domain\": {\"action\": \"call_agent\", \"config\": {\"agent_type\": \"domain-analyst\"}, \"next_step\": \"design_architecture\"}, \"design_architecture\": {\"action\": \"call_agent\", \"config\": {\"agent_type\": \"site-architect\"}, \"next_step\": \"create_content\"}, \"create_content\": {\"action\": \"call_agent\", \"config\": {\"agent_type\": \"content-creator\"}, \"next_step\": \"develop_site\"}, \"develop_site\": {\"action\": \"call_agent\", \"config\": {\"agent_type\": \"html-developer\"}, \"next_step\": \"publish_site\"}, \"publish_site\": {\"action\": \"call_agent\", \"config\": {\"agent_type\": \"site-publisher\"}, \"next_step\": \"complete\"}, \"complete\": {\"action\": \"complete_workflow\"}}}}'::jsonb),

-- Domain Analyst
('domain-analyst', 'Domain Analyst', 'Analyzes business domains', 'analyzer',
'[\"domain_analysis\", \"market_research\", \"competitor_analysis\"]'::jsonb,
'{\"process\": \"system.agent.domain-analyst.process\", \"response\": \"system.responses.domain-analyst\", \"error\": \"system.errors.domain-analyst\", \"dlq\": \"dlq.domain-analyst\"}'::jsonb,
'{\"workflow\": {\"start_step\": \"analyze\", \"steps\": {\"analyze\": {\"action\": \"execute_llm_prompt\", \"next_step\": \"complete\"}, \"complete\": {\"action\": \"complete_workflow\"}}}}'::jsonb),

-- Site Architect  
('site-architect', 'Site Architect', 'Designs website architecture', 'designer',
'[\"site_architecture\", \"information_architecture\", \"ux_design\"]'::jsonb,
'{\"process\": \"system.agent.site-architect.process\", \"response\": \"system.responses.site-architect\", \"error\": \"system.errors.site-architect\", \"dlq\": \"dlq.site-architect\"}'::jsonb,
'{\"workflow\": {\"start_step\": \"design\", \"steps\": {\"design\": {\"action\": \"execute_llm_prompt\", \"next_step\": \"complete\"}, \"complete\": {\"action\": \"complete_workflow\"}}}}'::jsonb),

-- Content Creator
('content-creator', 'Content Creator', 'Creates website content', 'creator',
'[\"content_creation\", \"copywriting\", \"seo_optimization\"]'::jsonb,
'{\"process\": \"system.agent.content-creator.process\", \"response\": \"system.responses.content-creator\", \"error\": \"system.errors.content-creator\", \"dlq\": \"dlq.content-creator\"}'::jsonb,
'{\"workflow\": {\"start_step\": \"create\", \"steps\": {\"create\": {\"action\": \"execute_llm_prompt\", \"next_step\": \"complete\"}, \"complete\": {\"action\": \"complete_workflow\"}}}}'::jsonb),

-- Content Researcher
('content-researcher', 'Content Researcher', 'Researches content topics', 'researcher',
'[\"content_research\", \"fact_checking\", \"source_gathering\"]'::jsonb,
'{\"process\": \"system.agent.content-researcher.process\", \"response\": \"system.responses.content-researcher\", \"error\": \"system.errors.content-researcher\", \"dlq\": \"dlq.content-researcher\"}'::jsonb,
'{\"workflow\": {\"start_step\": \"research\", \"steps\": {\"research\": {\"action\": \"execute_llm_prompt\", \"next_step\": \"complete\"}, \"complete\": {\"action\": \"complete_workflow\"}}}}'::jsonb),

-- Visual Designer
('visual-designer', 'Visual Designer', 'Creates visual designs', 'designer',
'[\"visual_design\", \"ui_design\", \"branding\"]'::jsonb,
'{\"process\": \"system.agent.visual-designer.process\", \"response\": \"system.responses.visual-designer\", \"error\": \"system.errors.visual-designer\", \"dlq\": \"dlq.visual-designer\"}'::jsonb,
'{\"workflow\": {\"start_step\": \"design\", \"steps\": {\"design\": {\"action\": \"execute_llm_prompt\", \"next_step\": \"complete\"}, \"complete\": {\"action\": \"complete_workflow\"}}}}'::jsonb),

-- HTML Developer
('html-developer', 'HTML Developer', 'Develops HTML/CSS/JS', 'developer',
'[\"html_development\", \"css_styling\", \"javascript_development\"]'::jsonb,
'{\"process\": \"system.agent.html-developer.process\", \"response\": \"system.responses.html-developer\", \"error\": \"system.errors.html-developer\", \"dlq\": \"dlq.html-developer\"}'::jsonb,
'{\"workflow\": {\"start_step\": \"develop\", \"steps\": {\"develop\": {\"action\": \"execute_llm_prompt\", \"next_step\": \"complete\"}, \"complete\": {\"action\": \"complete_workflow\"}}}}'::jsonb),

-- Site Publisher
('site-publisher', 'Site Publisher', 'Publishes sites to hosting', 'publisher',
'[\"file_upload\", \"s3_storage\", \"website_publishing\"]'::jsonb,
'{\"process\": \"system.agent.site-publisher.process\", \"response\": \"system.responses.site-publisher\", \"error\": \"system.errors.site-publisher\", \"dlq\": \"dlq.site-publisher\"}'::jsonb,
'{\"workflow\": {\"start_step\": \"validate\", \"steps\": {\"validate\": {\"action\": \"validate_input\", \"next_step\": \"upload\"}, \"upload\": {\"action\": \"upload_to_s3\", \"next_step\": \"complete\"}, \"complete\": {\"action\": \"complete_workflow\"}}}}'::jsonb)

ON CONFLICT (type) DO UPDATE SET
display_name = EXCLUDED.display_name,
description = EXCLUDED.description,
capabilities = EXCLUDED.capabilities,
topics = EXCLUDED.topics,
default_config = EXCLUDED.default_config,
updated_at = NOW();"

--

kubectl -n ai-persona-system exec -it postgres-templates-0 -- psql -U templates_user -d templates_db -c "
UPDATE agent_definitions
SET
env_vars = '[
{\"name\": \"ENABLE_LOCAL_ACTIONS\", \"value\": \"true\"},
{\"name\": \"LOCAL_ACTION_MODULES\", \"value\": \"storage_actions\"}
]'::jsonb,
resources = '{
\"requests\": {\"cpu\": \"200m\", \"memory\": \"512Mi\"},
\"limits\": {\"cpu\": \"1000m\", \"memory\": \"2Gi\"}
}'::jsonb
WHERE type = 'site-publisher';"

$ kubectl -n ai-persona-system exec -it postgres-clients-0 -- psql -U clients_user -d clients_db -c "
UPDATE agent_definitions
SET
env_vars = '[
{\"name\": \"ENABLE_LOCAL_ACTIONS\", \"value\": \"true\"},
{\"name\": \"LOCAL_ACTION_MODULES\", \"value\": \"storage_actions\"}
]'::jsonb,
resources = '{
\"requests\": {\"cpu\": \"200m\", \"memory\": \"512Mi\"},
\"limits\": {\"cpu\": \"1000m\", \"memory\": \"2Gi\"}
}'::jsonb,
is_active = true
WHERE type = 'site-publisher';"
UPDATE 1


--

kubectl -n ai-persona-system exec -it postgres-clients-0 -- psql -U clients_user -d clients_db -c "
UPDATE client_demo_client.agent_instances
SET config = jsonb_set(
config,
'{workflow}',
jsonb_build_object(
'start_step', 'process_request',
'steps', jsonb_build_object(
'process_request', jsonb_build_object(
'action', 'call_agent',
'description', 'Call the requested agent',
'next_step', 'complete'
),
'complete', jsonb_build_object(
'action', 'complete_workflow',
'description', 'Complete the workflow'
)
)
)
)
WHERE id = '00000000-0000-0000-0000-000000000001';"