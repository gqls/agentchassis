-- ============================================================================
-- VERSION 2 AGENTS - Unified Site Builder Architecture
-- ============================================================================
-- These are v2 agents that work alongside existing v1 agents.
-- v1 agents continue to work as before.
-- v2 agents use the new pages/components structure.
--
-- To use v2: reference agent_type with version, or update workflows to use v2
-- ============================================================================

-- ============================================================================
-- 3. MULTIPAGE-WEBSITE-BUILDER-V2 - Uses pages array loop
-- ============================================================================
INSERT INTO agent_definitions (
    id,
    type,
    display_name,
    description,
    category,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    version,
    input_contract,
    output_contract
)
SELECT
    gen_random_uuid(),
    'multipage-website-builder',
    'Multipage Website Builder V2',
    'Builds websites using unified pages/components architecture (v2)',
    category,
    jsonb_build_object(
            'workflow', jsonb_build_object(
            'start_step', 'spawn_strategist',
            'steps', jsonb_build_object(
                    'spawn_strategist', jsonb_build_object(
                            'action', 'spawn_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'chief-strategist',
                                    'agent_version', 2,
                                    'role', 'strategist'
                                      ),
                            'next_step', 'spawn_content_creator',
                            'output_field', 'strategist_info',
                            'description', 'Spawn v2 strategist'
                                        ),
                    'spawn_content_creator', jsonb_build_object(
                            'action', 'spawn_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'content-creator',
                                    'agent_version', 2,
                                    'role', 'writer'
                                      ),
                            'next_step', 'spawn_html_developer',
                            'output_field', 'writer_info',
                            'description', 'Spawn v2 content creator'
                                             ),
                    'spawn_html_developer', jsonb_build_object(
                            'action', 'spawn_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'html-developer',
                                    'role', 'developer'
                                      ),
                            'next_step', 'spawn_deployer',
                            'output_field', 'developer_info',
                            'description', 'Spawn HTML developer'
                                            ),
                    'spawn_deployer', jsonb_build_object(
                            'action', 'spawn_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'deployer-agent',
                                    'role', 'deployer'
                                      ),
                            'next_step', 'call_strategist',
                            'output_field', 'deployer_info',
                            'description', 'Spawn deployer'
                                      ),
                    'call_strategist', jsonb_build_object(
                            'action', 'call_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'chief-strategist',
                                    'target_role', 'strategist',
                                    'timeout_seconds', 120
                                      ),
                            'next_step', 'generate_pages_loop',
                            'output_field', 'page_plan',
                            'description', 'Get page plan from v2 strategist'
                                       ),
                    'generate_pages_loop', jsonb_build_object(
                            'action', 'loop',
                            'config', jsonb_build_object(
                                    'iterate_over', 'page_plan.plan_data.pages',
                                    'loop_var', 'current_page',
                                    'max_iterations', 10,
                                    'substeps', jsonb_build_object(
                                            'generate_content', jsonb_build_object(
                                                    'action', 'call_agent',
                                                    'config', jsonb_build_object(
                                                            'agent_type', 'content-creator',
                                                            'target_role', 'writer',
                                                            'input_fields', jsonb_build_array('current_page', 'input_data', 'page_plan'),
                                                            'timeout_seconds', 180
                                                              ),
                                                    'next_step', 'create_html',
                                                    'output_field', 'page_content',
                                                    'description', 'Generate content for page'
                                                                ),
                                            'create_html', jsonb_build_object(
                                                    'action', 'call_agent',
                                                    'config', jsonb_build_object(
                                                            'agent_type', 'html-developer',
                                                            'target_role', 'developer',
                                                            'input_fields', jsonb_build_array('page_content', 'current_page', 'input_data', 'page_plan'),
                                                            'timeout_seconds', 180
                                                              ),
                                                    'output_field', 'page_html',
                                                    'description', 'Convert content to HTML'
                                                           )
                                                )
                                      ),
                            'next_step', 'assemble_site',
                            'output_field', 'all_pages',
                            'description', 'Generate all pages'
                                           ),
                    'assemble_site', jsonb_build_object(
                            'action', 'assemble_multipage_site',
                            'config', jsonb_build_object(
                                    'pages_field', 'all_pages',
                                    'add_navigation', true,
                                    'generate_standard_pages', false
                                      ),
                            'next_step', 'deploy',
                            'output_field', 'site_files',
                            'description', 'Assemble pages with navigation'
                                     ),
                    'deploy', jsonb_build_object(
                            'action', 'call_agent',
                            'config', jsonb_build_object(
                                    'agent_type', 'deployer-agent',
                                    'target_role', 'deployer',
                                    'input_fields', jsonb_build_array('site_files', 'input_data'),
                                    'timeout_seconds', 180
                                      ),
                            'next_step', 'complete',
                            'output_field', 'deployment_result',
                            'description', 'Deploy to repository'
                              ),
                    'complete', jsonb_build_object(
                            'action', 'complete_workflow',
                            'description', 'Build complete'
                                )
                     )
                        ),
            'processing_mode', 'orchestration',
            'timeout_seconds', 900
    ),
    true,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config,
    2,  -- VERSION 2
    '{"required": ["input_data"], "expects": {"input_data.domain": "string", "input_data.objective": "string"}}'::jsonb,
    '{"produces": "deployment_result", "format": {"type": "object"}}'::jsonb
FROM agent_definitions
WHERE type = 'multipage-website-builder' AND version = 1
    ON CONFLICT (type, version) DO UPDATE SET
    default_config = EXCLUDED.default_config,
                                       description = EXCLUDED.description,
                                       display_name = EXCLUDED.display_name,
                                       input_contract = EXCLUDED.input_contract,
                                       output_contract = EXCLUDED.output_contract,
                                       updated_at = NOW();


-- ============================================================================
-- VERIFICATION
-- ============================================================================
SELECT
    type,
    version,
    display_name,
    substring(description from 1 for 60) as desc_preview
FROM agent_definitions
WHERE type IN ('chief-strategist', 'content-creator', 'multipage-website-builder')
ORDER BY type, version;