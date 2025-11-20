-- ============================================================================
-- HITL Content Approval Agent Group Definition
-- For EBORG - Evidence-Based Organisational Planning
-- ============================================================================

INSERT INTO agent_group_definitions (
    id,
    name,
    group_type,
    agent_configs,
    orchestration_workflow,
    version,
    created_at,
    updated_at
) VALUES (
             gen_random_uuid(),
             'Content Approval with HITL',
             'content-approval-hitl',
             jsonb_build_array(
                     jsonb_build_object('role', 'content_writer', 'agent_type', 'simple-content-writer-with-approval')
             ),
             jsonb_build_object(
                     'start_step', 'spawn_content_writer',
                     'steps', jsonb_build_object(
                             'spawn_content_writer', jsonb_build_object(
                             'action', 'spawn_agent',
                             'config', jsonb_build_object(
                                     'role', 'content_writer',
                                     'agent_type', 'simple-content-writer-with-approval'
                                       ),
                             'description', 'Spawn content writer agent with approval workflow',
                             'next_step', 'generate_content'
                                                     ),
                             'generate_content', jsonb_build_object(
                                     'action', 'call_agent',
                                     'config', jsonb_build_object(
                                             'agent_type', 'simple-content-writer-with-approval',
                                             'target_role', 'content_writer',
                                             'input_data', jsonb_build_object(
                                                     'business_name', '{{.input_data.business_name}}',
                                                     'business_type', '{{.input_data.business_type}}',
                                                     'business_description', '{{.input_data.business_description}}'
                                                           )
                                               ),
                                     'description', 'Generate content and await approval',
                                     'next_step', 'aggregate_results'
                                                 ),
                             'aggregate_results', jsonb_build_object(
                                     'action', 'aggregate_data',
                                     'config', jsonb_build_object(
                                             'response_fields', jsonb_build_array('generate_content'),
                                             'strategy', 'merge_responses'
                                               ),
                                     'description', 'Collect approved content',
                                     'next_step', 'complete'
                                                  ),
                             'complete', jsonb_build_object(
                                     'action', 'complete_workflow',
                                     'description', 'Return approved content'
                                         )
                              )
             ),
             1,  -- version
             now(),
             now()
         ) ON CONFLICT (group_type)
DO UPDATE SET
    name = EXCLUDED.name,
                  agent_configs = EXCLUDED.agent_configs,
                  orchestration_workflow = EXCLUDED.orchestration_workflow,
                  version = EXCLUDED.version,
                  updated_at = now()
                  RETURNING id, name, group_type, version;

-- Verify the group was created
SELECT
    group_type,
    name,
    version,
    jsonb_object_keys(orchestration_workflow->'steps') as steps
FROM agent_group_definitions
WHERE group_type = 'content-approval-hitl';