-- test/migrations/002_test_data.sql
-- Insert test data for agent spawning tests

-- Insert test agent groups
INSERT INTO agent_groups (id, name, group_type, agent_configs, orchestration_workflow)
VALUES
-- Website builder group
('11111111-1111-1111-1111-111111111111',
 'Website Builder Team',
 'website-builder',
 '[
    {"role": "architect", "agent_type": "site-architect"},
    {"role": "designer", "agent_type": "visual-designer"},
    {"role": "developer", "agent_type": "html-developer"},
    {"role": "publisher", "agent_type": "site-publisher"}
 ]'::jsonb,
 '{
    "start_step": "plan",
    "steps": {
        "plan": {"action": "create_site_plan", "next_step": "design"},
        "design": {"action": "create_design", "next_step": "develop"},
        "develop": {"action": "build_html", "next_step": "publish"},
        "publish": {"action": "publish_site", "next_step": "complete"},
        "complete": {"action": "complete_workflow"}
    }
 }'::jsonb),

-- Content creation group
('22222222-2222-2222-2222-222222222222',
 'Content Creation Team',
 'content-team',
 '[
    {"role": "researcher", "agent_type": "researcher"},
    {"role": "writer", "agent_type": "content-creator"},
    {"role": "editor", "agent_type": "editor"}
 ]'::jsonb,
 '{
    "start_step": "research",
    "steps": {
        "research": {"action": "research_topic", "next_step": "write"},
        "write": {"action": "create_content", "next_step": "edit"},
        "edit": {"action": "edit_content", "next_step": "complete"},
        "complete": {"action": "complete_workflow"}
    }
 }'::jsonb),

-- Simple test group
('33333333-3333-3333-3333-333333333333',
 'Test Team',
 'test-group',
 '[
    {"role": "worker", "agent_type": "generic"}
 ]'::jsonb,
 '{
    "start_step": "process",
    "steps": {
        "process": {"action": "process_task", "next_step": "complete"},
        "complete": {"action": "complete_workflow"}
    }
 }'::jsonb)

    ON CONFLICT (id) DO NOTHING;

-- Insert test agent instances
INSERT INTO client_demo_client.agent_instances (id, template_id, owner_user_id, name, config, is_active)
VALUES
    ('00000000-0000-0000-0000-000000000001',
     '2a540b98-85d5-4762-a692-538bcf1be395',
     'test_user',
     'test-generic-agent',
     '{
        "agent_type": "generic",
        "topic": "system.agent.generic.requests",
        "capabilities": ["testing"],
        "workflow": {
            "start_step": "spawn_website_team",
            "steps": {
                "spawn_website_team": {
                    "action": "spawn_group",
                    "config": {
                        "group_type": "website-builder"
                    },
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow"
                }
            }
        }
     }'::jsonb,
     true)
    ON CONFLICT (id) DO NOTHING;