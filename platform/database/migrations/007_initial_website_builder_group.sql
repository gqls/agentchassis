-- Create the initial website builder group
-- In 007_initial_website_builder_group.sql
INSERT INTO agent_groups (
    id,
    name,
    group_type,
    version,
    agent_configs,
    orchestration_workflow,
    capabilities
) VALUES (
             '00000000-0000-0000-0000-000000001000',
             'Website Builder Team v1',
             'website-builder',
             '1.0.0',
             '[
                 {"role": "orchestrator", "agent_type": "website-builder"},
                 {"role": "domain_analyst", "agent_type": "domain-analyst"},
                 {"role": "site_architect", "agent_type": "site-architect"},
                 {"role": "content_researcher", "agent_type": "content-creator"},
                 {"role": "html_developer", "agent_type": "html-developer"},
                 {"role": "visual_designer", "agent_type": "visual-designer"},
                 {"role": "site_publisher", "agent_type": "site-publisher"}
             ]'::jsonb,
             '{
                 "start_step": "validate_request",
                 "steps": {
                     "validate_request": {
                         "action": "validate_input",
                         "next_step": "analyze_domain"
                     },
                     "analyze_domain": {
                         "action": "call_agent",
                         "config": {"agent_type": "domain-analyst"},
                         "next_step": "architect_site"
                     },
                     "architect_site": {
                         "action": "call_agent",
                         "config": {"agent_type": "site-architect"},
                         "next_step": "gather_content"
                     },
                     "gather_content": {
                         "action": "call_agent",
                         "config": {"agent_type": "content-creator"},
                         "next_step": "create_visuals"
                     },
                     "create_visuals": {
                         "action": "call_agent",
                         "config": {"agent_type": "visual-designer"},
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
             '["website-creation", "html", "css", "design", "publishing"]'::jsonb
         )
    ON CONFLICT (id) DO UPDATE SET
    orchestration_workflow = EXCLUDED.orchestration_workflow,
                            updated_at = NOW();