-- deployer-agent
-- Handles git operations and deployment
UPDATE agent_definitions
SET input_contract = '{
    "required": ["site_record"],
    "optional": ["page", "assembled_page", "pages_built", "site_files"]
}'::jsonb,
    output_contract = '{
    "produces": ["commit_sha", "deploy_url", "deployment_status"]
}'::jsonb
WHERE type = 'deployer-agent';

