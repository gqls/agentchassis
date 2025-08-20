-- 007_initial_website_builder_group.sql
-- Creates the initial website builder agent group configuration

INSERT INTO agent_groups (
    id,
    name,
    description,
    group_type,
    version,
    agent_configs,
    orchestration_workflow,
    capabilities,
    tags
) VALUES (
             '00000000-0000-0000-0000-000000001000'::uuid,
             'Website Builder Team v1',
             'Complete website creation team with domain analysis, architecture, content, development, and publishing capabilities',
             'website-builder',
             '1.0.0',
             '[
                 {"role": "orchestrator", "agent_type": "website-builder", "required": true},
                 {"role": "domain_analyst", "agent_type": "domain-analyst", "required": true},
                 {"role": "site_architect", "agent_type": "site-architect", "required": true},
                 {"role": "content_researcher", "agent_type": "content-researcher", "required": true},
                 {"role": "content_writer", "agent_type": "content-creator", "required": true},
                 {"role": "html_developer", "agent_type": "html-developer", "required": true},
                 {"role": "visual_designer", "agent_type": "visual-designer", "required": false},
                 {"role": "site_publisher", "agent_type": "site-publisher", "required": true}
             ]'::jsonb,
             '{
                 "start_step": "validate_request",
                 "steps": {
                     "validate_request": {
                         "action": "validate_input",
                         "description": "Validate the incoming website creation request",
                         "next_step": "analyze_domain"
                     },
                     "analyze_domain": {
                         "action": "call_agent",
                         "config": {"agent_type": "domain-analyst"},
                         "description": "Analyze the domain and business type",
                         "next_step": "architect_site"
                     },
                     "architect_site": {
                         "action": "call_agent",
                         "config": {"agent_type": "site-architect"},
                         "description": "Design site structure and navigation",
                         "next_step": "gather_content"
                     },
                     "gather_content": {
                         "action": "parallel_agents",
                         "config": {
                             "agents": [
                                 {"agent_type": "content-researcher", "role": "research"},
                                 {"agent_type": "content-creator", "role": "write"}
                             ]
                         },
                         "description": "Research and create content for the website",
                         "next_step": "create_visuals"
                     },
                     "create_visuals": {
                         "action": "call_agent",
                         "config": {"agent_type": "visual-designer"},
                         "description": "Generate logos and visual assets",
                         "next_step": "develop_site"
                     },
                     "develop_site": {
                         "action": "call_agent",
                         "config": {"agent_type": "html-developer"},
                         "description": "Generate HTML/CSS/JS code for all pages",
                         "next_step": "publish_site"
                     },
                     "publish_site": {
                         "action": "call_agent",
                         "config": {"agent_type": "site-publisher"},
                         "description": "Deploy website to storage bucket",
                         "next_step": "complete"
                     },
                     "complete": {
                         "action": "complete_workflow",
                         "description": "Finalize the website creation process"
                     }
                 }
             }'::jsonb,
             '["website-creation", "html", "css", "design", "publishing", "content-generation", "seo"]'::jsonb,
             '["website", "builder", "automated", "ai-powered"]'::jsonb
         )
    ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
                            description = EXCLUDED.description,
                            agent_configs = EXCLUDED.agent_configs,
                            orchestration_workflow = EXCLUDED.orchestration_workflow,
                            capabilities = EXCLUDED.capabilities,
                            tags = EXCLUDED.tags,
                            updated_at = NOW();