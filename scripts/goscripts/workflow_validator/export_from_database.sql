-- Export single agent
SELECT json_build_object(
               'id', id::text,
               'type', type,
               'display_name', display_name,
               'default_config', default_config,
               'input_contract', input_contract,
               'output_contract', output_contract
       ) FROM agent_definitions WHERE type = 'pageflow-builder';

-- Export multiple agents
SELECT json_agg(json_build_object(
        'id', id::text,
        'type', type,
        'display_name', display_name,
        'default_config', default_config,
        'input_contract', input_contract,
        'output_contract', output_contract
                )) FROM agent_definitions WHERE type IN ('pageflow-builder', 'page-content-writer', 'deployer-agent');