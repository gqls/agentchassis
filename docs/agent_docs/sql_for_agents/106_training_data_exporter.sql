-- ============================================================================
-- Create training-data-exporter agent definition
-- ============================================================================
-- A single-action deterministic agent that exports successful LLM calls from
-- llm_call_log to a training-data JSONL file in ChatML + metadata format.
--
-- Trigger via Kafka on system.agent.generic.requests, same pattern as
-- rag-test-agent. Input payload:
--
--   {
--     "action": "orchestrate",
--     "config": {"agent_type": "training-data-exporter"},
--     "input_data": {
--       "agent_type":           "page-content-writer",
--       "step_name":            "process_sections_loop_iter_0_generate_content",
--       "model_filter":         "claude-sonnet-4-6",
--       "output_path":          "/tmp/training_exports/page_content_writer_iter0.jsonl",
--       "include_fenced":       true,
--       "strict_json":          true,
--       "min_response_length":  10,
--       "max_rows":             100000
--     }
--   }
--
-- After it runs, retrieve the file from the chassis pod:
--   kubectl -n ai-persona-system cp \
--     <chassis-pod>:/tmp/training_exports/page_content_writer_iter0.jsonl \
--     ./page_content_writer_iter0.jsonl
-- ============================================================================

INSERT INTO agent_definitions (
    type, display_name, description, category, default_config, is_active
) VALUES (
             'training-data-exporter',
             'Training Data Exporter',
             'Exports successful LLM calls from llm_call_log as NDJSON training data in ChatML + metadata format. Configure target agent/step via input_data.',
             'experimental',
             '{
               "workflow": {
                 "start_step": "export",
                 "processing_mode": "orchestrator",
                 "timeout_seconds": 600,
                 "steps": {
                   "export": {
                     "action": "training_data_export",
                     "config": {
                       "agent_type":           "{{.input_data.agent_type}}",
                       "step_name":            "{{.input_data.step_name}}",
                       "model_filter":         "{{.input_data.model_filter}}",
                       "output_path":          "{{.input_data.output_path}}",
                       "include_fenced":       true,
                       "strict_json":          true,
                       "min_response_length":  10,
                       "max_rows":             100000
                     },
                     "next_step": "complete",
                     "output_field": "export_result",
                     "description": "Query llm_call_log, clean responses, write NDJSON"
                   },
                   "complete": {
                     "action": "complete_workflow",
                     "config": {
                       "output_fields": ["export_result"],
                       "success_message": "Training data export complete"
                     }
                   }
                 }
               }
             }'::jsonb,
             true
         )
    ON CONFLICT (type, version) WHERE deleted_at IS NULL DO UPDATE
                                                                SET default_config = EXCLUDED.default_config,
                                                                updated_at = NOW();

-- Verify
SELECT type, display_name, is_active, is_snapshot,
       jsonb_extract_path_text(default_config, 'workflow', 'start_step') as start_step
FROM agent_definitions
WHERE type = 'training-data-exporter' AND deleted_at IS NULL;

