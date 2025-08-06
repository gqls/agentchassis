-- Create the initial website builder group
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
             '{"entry_point": "website-builder"}'::jsonb,
             '["website-creation", "html", "css", "design", "publishing"]'::jsonb
         );