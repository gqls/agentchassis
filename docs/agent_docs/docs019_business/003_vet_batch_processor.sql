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

--

-- pass orchestration id through to verifier

-- vet_batch_processor_fixes.sql
--
-- Three fixes:
-- 1. Pass task_id through to verifier so tasks get completed by ID
-- 2. Reorder: spawn verifier FIRST, then loop batches until empty
-- 3. Add continue_on_error so individual failures don't kill the batch
--
-- New flow:
--   spawn_verifier → load_batch → check_batch
--     → (has items) process_batch → load_batch → check_batch
--     → (empty) complete_empty
--
-- Verifier is spawned once and reused for all batches.



-- Fix 2+3: Reorder steps and add continue_on_error
-- spawn_verifier becomes the start_step, its next_step → load_batch
-- check_batch then_step → process_batch (not spawn_verifier)
-- process_batch next_step → load_batch (loop back)
-- process_batch gets continue_on_error: true
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                jsonb_set(
                        jsonb_set(
                                jsonb_set(
                                        default_config,
                                        '{workflow,start_step}',
                                        '"spawn_verifier"'::jsonb
                                ),
                                '{workflow,steps,spawn_verifier,next_step}',
                                '"load_batch"'::jsonb
                        ),
                        '{workflow,steps,check_batch,config,then_step}',
                        '"process_batch"'::jsonb
                ),
                '{workflow,steps,process_batch,next_step}',
                '"load_batch"'::jsonb
        ),
        '{workflow,steps,process_batch,config,continue_on_error}',
        'true'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-batch-processor';


---

-- remove loop

-- vet_batch_processor_loop_fix.sql
--
-- Fix: remove loop-back pattern (loop steps can't re-expand).
-- Each batch processor run handles one batch and completes.
-- Pipeline runs again to drain the queue.
--
-- Changes:
--   process_batch.next_step: "load_batch" → "complete"
--   process_batch.max_iterations: 50 → 250
--   start_step stays at spawn_verifier (spawn once, process batch, complete)

-- 1. Fix the batch processor definition
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,process_batch,next_step}',
                '"complete"'::jsonb
        ),
        '{workflow,steps,process_batch,config,max_iterations}',
        '250'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'vet-batch-processor';

-- 2. Reset orphaned in_progress tasks from previous failed runs
UPDATE business_intel.collection_tasks
SET status = 'pending', started_at = NULL, orchestration_id = NULL
WHERE status = 'in_progress';

-- 3. Fail the stuck verifiers so they don't hold resources
UPDATE orchestration_states
SET status = 'FAILED',
    error = 'stuck at scrape_website - timeout goroutine lost',
    updated_at = NOW()
WHERE status = 'AWAITING_RESPONSES'
  AND owner_agent_type = 'vet-practice-verifier'
  AND updated_at < NOW() - INTERVAL '30 minutes';