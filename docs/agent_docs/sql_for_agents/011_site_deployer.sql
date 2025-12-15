-- Site Deployer
UPDATE agent_definitions
SET
    input_contract = '{
        "required": ["site_files", "input_data"],
        "expects": {
            "site_files": "object (map of files)",
            "input_data.domain": "string",
            "input_data.repo_name": "string"
        }
    }'::jsonb,
    output_contract = '{
        "produces": "deployment_result",
        "format": {
            "type": "object",
            "properties": {
                "status": "string",
                "url": "string",
                "commit_sha": "string"
            },
            "description": "Deployment status and URLs"
        }
    }'::jsonb
WHERE type = 'site-deployer';