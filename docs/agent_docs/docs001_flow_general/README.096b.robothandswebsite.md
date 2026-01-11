-- ============================================================================
-- Robot Hands Website Workflow
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
'Robot Hands Website',
'robot-hands-website',
jsonb_build_array(
jsonb_build_object('role', 'hero_writer', 'agent_type', 'content-creator-hero'),
jsonb_build_object('role', 'image_creator', 'agent_type', 'image-generator')
),
jsonb_build_object(
'start_step', 'spawn_hero_writer',
'steps', jsonb_build_object(
'spawn_hero_writer', jsonb_build_object(
'action', 'spawn_agent',
'config', jsonb_build_object('role', 'hero_writer', 'agent_type', 'content-creator-hero'),
'next_step', 'generate_hero'
),
'generate_hero', jsonb_build_object(
'action', 'call_agent',
'config', jsonb_build_object(
'agent_type', 'content-creator-hero',
'target_role', 'hero_writer',
'prompt', 'Write a compelling hero section for {{.business_name}}, a {{.business_type}}. Focus on precision robotics and automation.'
),
'next_step', 'generate_hero_image'
),
'generate_hero_image', jsonb_build_object(
'action', 'call_agent',
'config', jsonb_build_object(
'agent_type', 'image-generator',
'prompt', 'Professional photograph of precision robotic hands assembling electronic components, modern factory setting, dramatic lighting, photorealistic, 8k',
'width', 1920,
'height', 1080
),
'next_step', 'complete'
),
'complete', jsonb_build_object(
'action', 'complete_workflow'
)
)
),
1,
now(),
now()
) RETURNING id, name, group_type, version;

-- Verify
SELECT
group_type,
jsonb_object_keys(orchestration_workflow->'steps') as steps
FROM agent_group_definitions
WHERE group_type = 'robot-hands-website';



---

agent definition

-- ============================================================================
-- Create New Version of Image Generator Agent Type
-- ============================================================================
-- Keeps existing version, adds new v2 with orchestrator workflow
-- Messages can specify which version to use
-- ============================================================================
INSERT INTO agent_definitions (
id,
type,
display_name,
description,
category,
default_config,
is_active,
created_at,
updated_at,
deleted_at,
capabilities,
image_repository,
image_tag,
resources,
topics,
health_config,
env_vars,
version,
previous_version_id,
task_workflow,
delegation_preferences
)
VALUES (
gen_random_uuid(),
'image-generator',
'Image Generator',
'Creates images using AI generation with S3 storage (orchestrator mode)',
'adapter',

    -- default_config
    '{
      "processing_mode": "orchestrator",
      "model": "sdxl",
      "provider": "stability_ai",
      "api_config": {
        "base_url": "https://api.stability.ai/v1/generation/stable-diffusion-xl-1024-v1-0/text-to-image",
        "api_key_env_var": "IMAGE_API_KEY",
        "timeout_seconds": 60
      },
      "image_settings": {
        "default_width": 1920,
        "default_height": 1080,
        "default_format": "png",
        "max_width": 2048,
        "max_height": 2048
      },
      "capabilities": {
        "storage": {
          "enabled": true,
          "provider": "s3",
          "bucket_env_var": "IMAGE_BUCKET",
          "region_env_var": "S3_REGION",
          "endpoint_env_var": "S3_ENDPOINT",
          "access_key_env_var": "AWS_ACCESS_KEY_ID",
          "secret_key_env_var": "AWS_SECRET_ACCESS_KEY"
        }
      },
      "workflow": {
        "start_step": "generate_and_upload",
        "steps": {
          "generate_and_upload": {
            "action": "generate_image_and_upload",
            "description": "Generate image and upload to S3",
            "next_step": "complete"
          },
          "complete": {
            "action": "complete_workflow",
            "description": "Return image URI"
          }
        }
      }
    }'::jsonb,

    true,              -- is_active
    now(),
    now(),
    null,              -- deleted_at

    '["image-generation", "text-to-image", "s3-storage", "orchestration"]'::jsonb,

    'docker.io/aqls/agent-chassis',
    'v1.0.390',

    -- resources
    '{
      "requests": { "cpu": "100m", "memory": "256Mi" },
      "limits": { "cpu": "500m", "memory": "1Gi" }
    }'::jsonb,

    -- topics
    '{
      "dlq": "system.agent.image-generator.dlq",
      "errors": "system.agent.image-generator.errors",
      "requests": "system.agent.image-generator.requests",
      "responses": "system.agent.image-generator.responses"
    }'::jsonb,

    -- health_config
    '{
      "port": 8080,
      "liveness_path": "/health",
      "readiness_path": "/ready",
      "initial_delay_seconds": 30
    }'::jsonb,

    -- env_vars
    '[
      {"name": "AWS_ACCESS_KEY_ID", "valueFrom": {"secretKeyRef": {"name": "personae-storage-secrets", "key": "AWS_ACCESS_KEY_ID"}}},
      {"name": "AWS_SECRET_ACCESS_KEY", "valueFrom": {"secretKeyRef": {"name": "personae-storage-secrets", "key": "AWS_SECRET_ACCESS_KEY"}}},
      {"name": "S3_ENDPOINT", "valueFrom": {"configMapKeyRef": {"name": "storage-config", "key": "S3-ENDPOINT"}}},
      {"name": "S3_REGION", "valueFrom": {"configMapKeyRef": {"name": "storage-config", "key": "S3-REGION"}}},
      {"name": "IMAGE_BUCKET", "valueFrom": {"configMapKeyRef": {"name": "storage-config", "key": "image_bucket"}}},
      {"name": "IMAGE_API_KEY", "valueFrom": {"secretKeyRef": {"name": "image-api-credentials", "key": "api-key"}}}
    ]'::jsonb,

    2,             -- version (example static)
    null,          -- previous_version_id

    -- task_workflow
    '{
      "start_step": "generate_and_upload",
      "steps": {
        "generate_and_upload": {
          "action": "generate_image_and_upload",
          "description": "Generate image and upload to S3",
          "next_step": "complete"
        },
        "complete": {
          "action": "complete_workflow",
          "description": "Return image URI"
        }
      }
    }'::jsonb,

    -- delegation_preferences
    '{"prefer_delegation": true, "fallback_to_self": true}'::jsonb
);


---


allow versions into db
-- 1. Drop the current unique constraint that only allows one version per type
ALTER TABLE agent_definitions
DROP CONSTRAINT IF EXISTS agent_definitions_type_key;

-- 2. Add a new unique constraint on type + version
ALTER TABLE agent_definitions
ADD CONSTRAINT agent_definitions_type_version_key UNIQUE (type, version);

ALTER TABLE agent_capabilities      DROP CONSTRAINT IF EXISTS agent_capabilities_agent_type_fkey;
ALTER TABLE agent_dependencies      DROP CONSTRAINT IF EXISTS agent_dependencies_agent_type_fkey;
ALTER TABLE agent_metrics_config    DROP CONSTRAINT IF EXISTS agent_metrics_config_agent_type_fkey;
ALTER TABLE agent_default_configs   DROP CONSTRAINT IF EXISTS agent_default_configs_agent_type_fkey;
ALTER TABLE agent_group_members     DROP CONSTRAINT IF EXISTS agent_group_members_agent_type_fkey;

ALTER TABLE agent_capabilities
ADD CONSTRAINT agent_capabilities_agent_type_fkey
FOREIGN KEY (agent_type) REFERENCES agent_definitions(type);

ALTER TABLE agent_dependencies
ADD CONSTRAINT agent_dependencies_agent_type_fkey
FOREIGN KEY (agent_type) REFERENCES agent_definitions(type);

ALTER TABLE agent_metrics_config
ADD CONSTRAINT agent_metrics_config_agent_type_fkey
FOREIGN KEY (agent_type) REFERENCES agent_definitions(type);

ALTER TABLE agent_default_configs
ADD CONSTRAINT agent_default_configs_agent_type_fkey
FOREIGN KEY (agent_type) REFERENCES agent_definitions(type);

ALTER TABLE agent_group_members
ADD CONSTRAINT agent_group_members_agent_type_fkey
FOREIGN KEY (agent_type) REFERENCES agent_definitions(type);
