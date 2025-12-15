-- View workflow with contracts to manually validate compatibility
CREATE OR REPLACE VIEW workflow_contract_chain AS
SELECT
    ad.type,
    ad.display_name,
    ad.input_contract,
    ad.output_contract,
    jsonb_pretty(ad.default_config->'workflow') as workflow
FROM agent_definitions ad
WHERE ad.type IN (
                  'website-builder',
                  'domain-analyst',
                  'site-architect',
                  'content-creator',
                  'html-developer',
                  'multipage-wrapper',
                  'site-deployer'
    )
ORDER BY ad.type;

-- Check it out
SELECT * FROM workflow_contract_chain;

===============
---------------

Step 4: Quick Contract Checker Query
Use this to manually verify a workflow_s contract chain:

-- Check website-builder workflow contracts
WITH workflow_steps AS (
    SELECT
        jsonb_object_keys(default_config->'workflow'->'steps') as step_name,
        default_config->'workflow'->'steps'->jsonb_object_keys(default_config->'workflow'->'steps') as step_config
    FROM agent_definitions
    WHERE type = 'website-builder'
),
step_agents AS (
    SELECT
        ws.step_name,
        ws.step_config->>'action' as action,
        ws.step_config->'config'->>'target_role' as target_role,
        ws.step_config->>'output_field' as output_field,
        ws.step_config->>'next_step' as next_step,
        ad.type as agent_type,
        ad.input_contract,
        ad.output_contract
    FROM workflow_steps ws
    LEFT JOIN agent_definitions ad ON ad.type = ws.step_config->'config'->>'agent_type'
    WHERE ws.step_config->>'action' IN ('call_agent', 'spawn_agent')
)
SELECT
    step_name,
    agent_type,
    output_field,
    output_contract->>'produces' as produces,
    next_step,
    input_contract->'required' as next_step_needs
FROM step_agents
ORDER BY step_name;

====

==========
==========



ALTER TABLE agent_definitions
ADD COLUMN input_contract jsonb,
ADD COLUMN output_contract jsonb;

-- Example:
UPDATE agent_definitions
SET
input_contract = '{
"required": ["site_architecture", "site_content"],
"optional": ["input_data"]
}'::jsonb,
output_contract = '{
"produces": "final_html",
"format": {
"type": "object",
"properties": {
"html": {"type": "string"},
"css": {"type": "string"}
}
}
}'::jsonb
WHERE type = 'html-developer';
```

This way you can:
1. **Validate workflows** - Check that agent A's output matches agent B's input requirements
2. **Document dependencies** - See what each agent needs without reading code
3. **Enable swapping** - Find alternative agents that satisfy the same contract

## Named Groups as Organization, Not Coupling

For Option 3, think of groups as **workflow templates** rather than tight coupling:
```
"html-pipeline" group = {
workflow: [content-creator → html-developer → multipage-wrapper],
contract: {
input: domain_analysis + site_architecture,
output: site_files
}
}

                                            ==

                                            -- Add contract columns to agent_definitions
ALTER TABLE agent_definitions
    ADD COLUMN IF NOT EXISTS input_contract jsonb,
    ADD COLUMN IF NOT EXISTS output_contract jsonb;

-- Add helpful comment
COMMENT ON COLUMN agent_definitions.input_contract IS 'Defines what data this agent expects as input';
COMMENT ON COLUMN agent_definitions.output_contract IS 'Defines what data this agent produces as output';

===
----
WITH workflow_steps AS (
    SELECT
        jsonb_object_keys(default_config->'workflow'->'steps') as step_name,
        default_config->'workflow'->'steps'->jsonb_object_keys(default_config->'workflow'->'steps') as step_config
    FROM agent_definitions
    WHERE type = 'website-builder'
),
step_details AS (
    SELECT
        ws.step_name,
        ws.step_config->>'action' as action,
        ws.step_config->'config'->>'target_role' as target_role,
        ws.step_config->'config'->>'agent_type' as config_agent_type,
        ws.step_config->>'output_field' as output_field,
        ws.step_config->>'next_step' as next_step
    FROM workflow_steps ws
),
step_with_contracts AS (
    SELECT
        sd.*,
        ad.type as resolved_agent_type,
        ad.input_contract,
        ad.output_contract
    FROM step_details sd
    LEFT JOIN agent_definitions ad ON (
        ad.type = sd.config_agent_type
        OR (sd.target_role = 'analyst' AND ad.type = 'domain-analyst')
        OR (sd.target_role = 'architect' AND ad.type = 'site-architect')
        OR (sd.target_role = 'content' AND ad.type = 'content-creator')
        OR (sd.target_role = 'developer' AND ad.type = 'html-developer')
        OR (sd.target_role = 'wrapper' AND ad.type = 'multipage-wrapper')
        OR (sd.target_role = 'deployer' AND ad.type = 'site-deployer')
    )
)
SELECT
    step_name,
    action,
    resolved_agent_type,
    output_field,
    output_contract->>'produces' as contract_produces,
    next_step,
    input_contract->'required' as needs
FROM step_with_contracts
ORDER BY
    CASE
    WHEN step_name LIKE 'spawn_%' THEN 1
    WHEN action = 'call_agent' THEN 2
    ELSE 3
END,
    step_name;