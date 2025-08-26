kubectl -n ai-persona-system exec -it postgres-clients-0 -- psql -U clients_user -d clients_db


-- Get all agent definitions with their workflows
SELECT
type,
display_name,
category,
default_config->'workflow' as workflow,
default_config->'workflow'->'steps' as workflow_steps
FROM agent_definitions
WHERE is_active = true
ORDER BY type;

-- Check specific problematic agents
SELECT
type,
jsonb_pretty(default_config->'workflow') as workflow
FROM agent_definitions
WHERE type IN ('domain-analyst', 'site-architect', 'content-creator', 'html-developer')
AND is_active = true;