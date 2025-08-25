-- test/migrations/002_test_data.sql
-- Insert test data for agent spawning tests

-- Insert test agent groups
INSERT INTO agent_groups (id, name, group_type, agent_configs, orchestration_workflow)
VALUES
-- Website builder group

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
    ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config, updated_at = NOW();


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
            "start_step": "validate",
            "steps": {
                "validate": {
                    "action": "validate_input",
                    "next_step": "transform"
                },
                "transform": {
                    "action": "transform_data",
                    "config": {"transformation": "uppercase"},
                    "next_step": "complete"
                },
                "complete": {
                    "action": "complete_workflow"
                }
            }
        }
     }'::jsonb,
     true)
    ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config, updated_at = NOW();


-- ADD THIS ENTIRE BLOCK TO THE END OF 002_test_data.sql

-- Create the specific instance of the generic agent for the demo client.
-- This is the agent that will receive the initial Kafka message.
INSERT INTO client_demo_client.agent_instances (id, template_id, owner_user_id, name, config, is_active)
VALUES
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'system_user', 'Default Generic Agent', '{
    "workflow": {
        "start_step": "spawn_website_team",
        "steps": {
            "spawn_website_team": {
                "action": "spawn_group",
                "config": { "group_type": "website-builder" },
                "next_step": "initiate_build"
            },
            "initiate_build": {
                "action": "start_orchestration",
                "next_step": "complete"
            },
            "complete": {
                "action": "complete_workflow"
            }
        }
    }
}'::jsonb, true)
    ON CONFLICT (id) DO UPDATE
                            SET
                                config = EXCLUDED.config,
                            updated_at = NOW();

