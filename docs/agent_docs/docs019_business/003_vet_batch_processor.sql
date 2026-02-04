-- =============================================
-- vet-batch-processor
-- =============================================
-- Loads a batch of pending collection tasks and processes them sequentially.
-- Spawns a vet-practice-verifier for each task.
-- Designed to run as a single pod, working through the queue.
--
-- Receives: { "batch_size": 10, "task_type": "initial_verification", "vertical_slug": "veterinary" }
-- Or can be triggered with no input to use defaults.

INSERT INTO agent_definitions (
    type, display_name, description, category,
    is_active, status, agent_category,
    capabilities, domain_tags,
    default_config,
    input_contract, output_contract
) VALUES (
             'vet-batch-processor',
             'Vet Batch Processor',
             'Loads pending verification tasks from the queue and processes them sequentially by spawning vet-practice-verifier agents. Designed for single-pod, low-throughput, polite data collection.',
             'orchestrator',
             TRUE, 'experimental', 'coordinator',
             '["batch-processing", "orchestration", "veterinary", "scheduling"]'::jsonb,
             '["veterinary", "business-intelligence", "batch-processing"]'::jsonb,
             '{
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 3600,
                 "ai_service": {
                     "provider": "anthropic",
                     "model": "claude-haiku-4-5",
                     "api_key_env_var": "ANTHROPIC_API_KEY"
                 },
                 "workflow": {
                     "start_step": "load_batch",
                     "steps": {
                         "load_batch": {
                             "action": "load_business_batch",
                             "description": "Claim next batch of pending tasks from the queue",
                             "config": {
                                 "batch_size": 10,
                                 "task_type": "initial_verification",
                                 "vertical_slug": "veterinary"
                             },
                             "output_field": "batch",
                             "next_step": "check_batch"
                         },

                         "check_batch": {
                             "action": "conditional",
                             "description": "Check if there are tasks to process",
                             "config": {
                                 "condition": "batch.batch_size > 0",
                                 "then_step": "spawn_verifier",
                                 "else_step": "complete_empty"
                             }
                         },

                         "spawn_verifier": {
                             "action": "spawn_agent",
                             "description": "Spawn a vet practice verifier agent",
                             "config": {
                                 "role": "verifier",
                                 "agent_type": "vet-practice-verifier"
                             },
                             "output_field": "verifier_agent",
                             "next_step": "process_batch"
                         },

                         "process_batch": {
                             "action": "loop",
                             "description": "Process each task in the batch sequentially",
                             "config": {
                                 "mode": "sequential",
                                 "items_field": "batch.items",
                                 "item_variable": "current_task",
                                 "max_iterations": 50,
                                 "sub_workflow": {
                                     "start_step": "call_verifier",
                                     "steps": {
                                         "call_verifier": {
                                             "action": "call_agent",
                                             "description": "Call verifier for this practice",
                                             "config": {
                                                 "agent_type": "vet-practice-verifier",
                                                 "target_role": "verifier",
                                                 "input_mapping": {
                                                     "business_id": "current_task.business_id",
                                                     "task_id": "current_task.task_id"
                                                 },
                                                 "timeout_seconds": 300
                                             },
                                             "output_field": "verifier_result",
                                             "next_step": "task_complete"
                                         },

                                         "task_complete": {
                                             "action": "loop_complete",
                                             "description": "Move to next task"
                                         }
                                     }
                                 }
                             },
                             "output_field": "batch_results",
                             "next_step": "complete"
                         },

                         "complete_empty": {
                             "action": "complete_workflow",
                             "description": "No pending tasks found",
                             "config": {
                                 "output_fields": ["batch"]
                             }
                         },

                         "complete": {
                             "action": "complete_workflow",
                             "description": "Batch processing complete",
                             "config": {
                                 "output_fields": ["batch", "batch_results"]
                             }
                         }
                     }
                 }
             }'::jsonb,
             '{
                 "required": [],
                 "optional": ["batch_size", "task_type", "vertical_slug"]
             }'::jsonb,
             '{
                 "produces": {
                     "batch": "object - the tasks that were claimed",
                     "batch_results": "object - results from processing each task"
                 }
             }'::jsonb
         ) ON CONFLICT (type, version) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    default_config = EXCLUDED.default_config,
    input_contract = EXCLUDED.input_contract,
    output_contract = EXCLUDED.output_contract,
    capabilities = EXCLUDED.capabilities,
    domain_tags = EXCLUDED.domain_tags,
    status = EXCLUDED.status,
    updated_at = NOW();

