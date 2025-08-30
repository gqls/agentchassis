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



---

// In the HTML developer agent's workflow (from database)
{
"workflow": {
"start_step": "generate_html",
"steps": {
"generate_html": {
"action": "execute_llm_prompt",
"config": {
"prompt_template": "html_generation",
"model": "claude-3-opus"
},
"next_step": "process_html"
},
"process_html": {
"action": "process_html",
"config": {
"add_meta_tags": true,
"optimize_images": true,
"minify": true,
"ensure_responsive": true,
"add_analytics": {
"enabled": false,
"ga_id": ""
}
},
"next_step": "validate_html"
},
"validate_html": {
"action": "validate_html",
"config": {
"strict": false,
"check_accessibility": true
},
"next_step": "store_if_valid"
},
"store_if_valid": {
"action": "conditional_branch",
"config": {
"condition": "validation.is_valid == true",
"true_branch": "store_output",
"false_branch": "complete"
}
},
"store_output": {
"action": "store_result",
"config": {
"content_field": "html",
"content_type": "text/html"
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow"
}
}
}
}

-- Update agent workflows to use route_storage for file storage
UPDATE agent_definitions
SET default_config = jsonb_set(
default_config,
'{workflow,steps,store_html}',
'{
"action": "route_storage",
"description": "Store HTML using configured storage backend",
"config": {
"storage_type": "s3",
"content_field": "final_html",
"content_type": "text/html",
"path_template": "websites/{{.ClientID}}/{{.CorrelationID}}/index.html",
"make_public": true
},
"next_step": "complete"
}'::jsonb
)
WHERE type = 'html-developer';


-- Add a workflow_mode column to agent_definitions
ALTER TABLE agent_definitions
ADD COLUMN task_workflow jsonb,
ADD COLUMN orchestrator_workflow jsonb;

-- Add workflow columns to agent_definitions
ALTER TABLE agent_definitions
ADD COLUMN orchestration_workflow jsonb,
ADD COLUMN task_workflow jsonb,
ADD COLUMN delegation_preferences jsonb DEFAULT '{"prefer_delegation": true, "fallback_to_self": true}';

-- Example: Update domain-analyst with both workflows
UPDATE agent_definitions
SET
-- When orchestrating, it can delegate analysis subtasks
orchestration_workflow = '{
"start_step": "check_complexity",
"steps": {
"check_complexity": {
"action": "evaluate_task",
"next_step": "decide_approach"
},
"decide_approach": {
"action": "conditional_route",
"config": {
"condition_field": "complexity",
"routes": {
"simple": "analyze_locally",
"complex": "delegate_analysis"
}
}
},
"delegate_analysis": {
"action": "call_agent",
"config": {
"agent_type": "researcher",
"fallback": "analyze_locally"
},
"next_step": "complete"
},
"analyze_locally": {
"action": "execute_llm_prompt",
"config": {
"prompt_template": "Analyze domain: {{.business_type}}"
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow"
}
}
}',
-- When called as a task processor
task_workflow = '{
"start_step": "analyze",
"steps": {
"analyze": {
"action": "execute_llm_prompt",
"config": {
"prompt_template": "Analyze domain: {{.business_type}}"
},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow"
}
}
}'
WHERE type = 'domain-analyst';

-- Keep the orchestrator workflow separate if needed
UPDATE agent_definitions
SET orchestrator_workflow = default_config->'workflow'
WHERE type = 'domain-analyst';

-- Update existing agents with task workflows
UPDATE agent_definitions
SET task_workflow = default_config->'workflow',
delegation_preferences = '{"prefer_delegation": false, "fallback_to_self": true}'::jsonb
WHERE default_config->'workflow' IS NOT NULL;

-- Set initial task workflows for existing agents
UPDATE agent_definitions
SET task_workflow = jsonb_build_object(
'start_step', 'execute',
'steps', jsonb_build_object(
'execute', jsonb_build_object(
'action', 'execute_llm_prompt',
'description', 'Execute task',
'next_step', 'complete'
),
'complete', jsonb_build_object(
'action', 'complete_workflow',
'description', 'Complete'
)
)
)
WHERE task_workflow IS NULL;


-- Insert the website-builder agent group
INSERT INTO agent_groups (id, name, group_type, agent_configs, orchestration_workflow, version, usage_count)
VALUES (
gen_random_uuid(),
'Website Builder Team',
'website-builder',
'[
{"role": "orchestrator", "agent_type": "website-builder"},
{"role": "domain_analyst", "agent_type": "domain-analyst"},
{"role": "site_architect", "agent_type": "site-architect"},
{"role": "content_researcher", "agent_type": "content-researcher"},
{"role": "content_writer", "agent_type": "content-creator"},
{"role": "visual_designer", "agent_type": "visual-designer"},
{"role": "html_developer", "agent_type": "html-developer"},
{"role": "site_publisher", "agent_type": "site-publisher"}
]'::jsonb,
'{
"start_step": "analyze_domain",
"steps": {
"analyze_domain": {
"action": "call_agent",
"config": {"agent_type": "domain-analyst"},
"next_step": "architect_site"
},
"architect_site": {
"action": "call_agent",
"config": {"agent_type": "site-architect"},
"next_step": "create_content"
},
"create_content": {
"action": "call_agent",
"config": {"agent_type": "content-creator"},
"next_step": "develop_site"
},
"develop_site": {
"action": "call_agent",
"config": {"agent_type": "html-developer"},
"next_step": "publish_site"
},
"publish_site": {
"action": "call_agent",
"config": {"agent_type": "site-publisher"},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow"
}
}
}'::jsonb,
1,
0
);


--
UPDATE agent_definitions
SET default_config = '{
"processing_mode": "orchestrator",
"workflow": {
"start_step": "spawn_agents",
"steps": {
"spawn_agents": {
"action": "spawn_group",
"description": "Spawn the website builder team",
"config": {},
"next_step": "start_orchestration"
},
"start_orchestration": {
"action": "start_orchestration",
"description": "Start orchestration",
"config": {"await_response": true},
"next_step": "complete"
},
"complete": {
"action": "complete_workflow",
"description": "Complete the workflow"
}
}
}
}'::jsonb
WHERE type = 'generic';

---

ALTER TABLE orchestration_states
ADD COLUMN parent_orchestration_id UUID,
-- ADD CONSTRAINT fk_parent_orchestration
FOREIGN KEY (parent_orchestration_id)
REFERENCES orchestration_states(orchestration_id)
ON DELETE CASCADE;

ALTER TABLE orchestration_states
DROP CONSTRAINT fk_parent_orchestration;

-- Index for performance
CREATE INDEX idx_parent_orchestration
ON orchestration_states(parent_orchestration_id)
WHERE parent_orchestration_id IS NOT NULL;



-- Verify current schema
\d orchestration_states

-- Should have at least:
-- orchestration_id UUID PRIMARY KEY
-- correlation_id UUID NOT NULL
-- owner_agent_id UUID NOT NULL  
-- parent_orchestration_id UUID (nullable)
-- client_id VARCHAR(100) NOT NULL
-- status VARCHAR(50)
-- current_step VARCHAR(100)
-- awaited_steps JSONB
-- collected_data JSONB
-- workflow_plan JSONB
-- ...


ALTER TABLE orchestration_states
ADD COLUMN fuel_budget INTEGER DEFAULT 1000;

ALTER TABLE orchestration_states
ALTER COLUMN owner_agent_id SET NOT NULL;


---

-- Add message deduplication table
CREATE TABLE IF NOT EXISTS processed_messages (
message_id UUID PRIMARY KEY,
correlation_id UUID NOT NULL,
orchestration_id UUID,
processed_at TIMESTAMP DEFAULT NOW(),
processed_by VARCHAR(100),
INDEX idx_correlation (correlation_id),
INDEX idx_processed_at (processed_at)
);

-- Add automatic cleanup for old processed messages (30 days)
CREATE INDEX IF NOT EXISTS idx_processed_messages_cleanup
ON processed_messages(processed_at);

-- Modify orchestration_states to add execution tracking
ALTER TABLE orchestration_states
ADD COLUMN IF NOT EXISTS currently_executing VARCHAR(100),
ADD COLUMN IF NOT EXISTS last_activity TIMESTAMP DEFAULT NOW(),
ADD COLUMN IF NOT EXISTS processing_node VARCHAR(100),
ADD COLUMN IF NOT EXISTS execution_started_at TIMESTAMP;

-- Update status enum (if using enum)
-- Otherwise just ensure these values are valid:
-- 'INITIALIZED', 'EXECUTING_STEP', 'AWAITING_RESPONSES', 'COMPLETED', 'FAILED'