-- ============================================================================
-- Simple Content Writer with HITL Approval Agent Definition
-- For EBORG - Evidence-Based Organisational Planning
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
    orchestrator_workflow,
    delegation_preferences
) VALUES (
             gen_random_uuid(),
             'simple-content-writer-with-approval',
             'Simple Content Writer with HITL Approval',
             'Generates content for organisations and waits for human approval before completion',
             'hitl-demo',

             -- default_config
             '{
                 "processing_mode": "task",
                 "ai_service": {
                     "provider": "anthropic",
                     "model": "claude-3-5-sonnet-20241022",
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "max_tokens": 500,
                 "temperature": 0.7,
                 "workflow": {
                     "start_step": "generate_draft",
                     "steps": {
                         "generate_draft": {
                             "action": "execute_llm_prompt",
                             "config": {
                                 "prompt_template": "Write a brief, professional description for {{.input_data.business_name}}. Context: {{.input_data.business_description}}. Focus on how they help organisations with {{.input_data.business_type}}. Keep it to 3-4 sentences that capture the essence of their value proposition.",
                                 "input_fields": ["input_data"]
                             },
                             "description": "Generate initial content draft",
                             "next_step": "await_human_approval"
                         },
                         "await_human_approval": {
                             "action": "await_approval",
                             "config": {
                                 "notification_data": {
                                     "type": "content_approval",
                                     "title": "Content Approval Required for {{.input_data.business_name}}",
                                     "description": "Please review and approve the generated organisational description",
                                     "content_field": "generate_draft",
                                     "metadata": {
                                         "business_name": "{{.input_data.business_name}}",
                                         "business_type": "{{.input_data.business_type}}"
                                     }
                                 },
                                 "timeout_seconds": 300,
                                 "include_generated_content": true
                             },
                             "description": "Wait for human approval of generated content",
                             "next_step": "process_approval"
                         },
                         "process_approval": {
                             "action": "process_data",
                             "config": {
                                 "input_fields": ["generate_draft", "await_human_approval"],
                                 "output_format": {
                                     "content": "{{.generate_draft.result}}",
                                     "approval_status": "{{.await_human_approval.approved}}",
                                     "approval_comments": "{{.await_human_approval.comments}}",
                                     "approved_by": "{{.await_human_approval.approved_by}}",
                                     "approved_at": "{{.await_human_approval.timestamp}}"
                                 }
                             },
                             "description": "Process approval response and prepare final output",
                             "next_step": "complete"
                         },
                         "complete": {
                             "action": "complete_workflow",
                             "description": "Return approved content with metadata"
                         }
                     }
                 }
             }'::jsonb,

             true,              -- is_active
             now(),
             now(),
             null,              -- deleted_at

             '["content-generation", "human-approval", "hitl", "organisational"]'::jsonb,

             'docker.io/aqls/agent-chassis',
             'v1.0.407',

             -- resources
             '{
                 "requests": {"cpu": "100m", "memory": "256Mi"},
                 "limits": {"cpu": "500m", "memory": "512Mi"}
             }'::jsonb,

             -- topics
             '{
                 "dlq": "system.agent.simple-content-writer-with-approval.dlq",
                 "errors": "system.agent.simple-content-writer-with-approval.errors",
                 "requests": "system.agent.simple-content-writer-with-approval.requests",
                 "responses": "system.agent.simple-content-writer-with-approval.responses"
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
                 {"name": "ANTHROPIC_API_KEY", "valueFrom": {"secretKeyRef": {"name": "anthropic-api-key", "key": "api-key"}}}
             ]'::jsonb,

             1,             -- version
             null,          -- previous_version_id

             -- task_workflow (same as in default_config)
             '{
                 "start_step": "generate_draft",
                 "steps": {
                     "generate_draft": {
                         "action": "execute_llm_prompt",
                         "config": {
                             "prompt_template": "Write a brief, professional description for {{.input_data.business_name}}. Context: {{.input_data.business_description}}. Focus on how they help organisations with {{.input_data.business_type}}. Keep it to 3-4 sentences that capture the essence of their value proposition.",
                             "input_fields": ["input_data"]
                         },
                         "description": "Generate initial content draft",
                         "next_step": "await_human_approval"
                     },
                     "await_human_approval": {
                         "action": "await_approval",
                         "config": {
                             "notification_data": {
                                 "type": "content_approval",
                                 "title": "Content Approval Required",
                                 "description": "Please review and approve the generated organisational description"
                             },
                             "timeout_seconds": 300,
                             "include_generated_content": true
                         },
                         "description": "Wait for human approval",
                         "next_step": "process_approval"
                     },
                     "process_approval": {
                         "action": "process_data",
                         "config": {
                             "input_fields": ["generate_draft", "await_human_approval"]
                         },
                         "description": "Process approval response",
                         "next_step": "complete"
                     },
                     "complete": {
                         "action": "complete_workflow",
                         "description": "Return approved content"
                     }
                 }
             }'::jsonb,

             null,          -- orchestrator_workflow (using task mode)

             -- delegation_preferences
             '{"prefer_delegation": false, "fallback_to_self": true}'::jsonb
         ) ON CONFLICT (type, version)
DO UPDATE SET
    display_name = EXCLUDED.display_name,
                  description = EXCLUDED.description,
                  default_config = EXCLUDED.default_config,
                  task_workflow = EXCLUDED.task_workflow,
                  updated_at = now()
                  RETURNING id, type, display_name, version;