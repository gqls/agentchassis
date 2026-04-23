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
---

-- ============================================================================
-- 002 — Agent definitions for training export v3
-- ============================================================================
-- Two agents:
--
--   training-data-exporter              (worker, specialist)
--      Does the real work. Reads llm_call_log, writes training_exports.
--      Must be spawned by a parent — no standalone invocation.
--
--   training-data-export-orchestrator   (wrapper, orchestrator)
--      Minimal three-step workflow: spawn → call → complete.
--      This is the agent you trigger via Kafka for manual exports.
--      Gives the worker a dedicated pod per doc 001 §"Every pod-running
--      agent needs a parent that spawned it".
--
-- Naming mirrors the canonical pattern in doc 001 (`med-export-orchestrator`
-- wrapping `med-json-exporter`).
-- ============================================================================

-- ----------------------------------------------------------------------------
-- WORKER: training-data-exporter
-- ----------------------------------------------------------------------------
-- Single-step workflow. Runs in a dedicated spawned pod (parent is the
-- orchestrator wrapper below). Parameters arrive via input_data per the
-- canonical call_agent pattern.
-- ----------------------------------------------------------------------------

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    default_config, is_active
) VALUES (
             'training-data-exporter',
             'Training Data Exporter',
             'Exports successful LLM calls from llm_call_log into training_exports schema (runs + rows tables) as ChatML + metadata records. Reads parameters from input_data (agent_type, step_name, model_filter, etc.). Must be invoked via training-data-export-orchestrator wrapper so it gets a dedicated pod.',
             'specialist',
             'specialist',
             'experimental',
             '{
               "workflow": {
                 "start_step": "export",
                 "steps": {
                   "export": {
                     "action": "training_data_export",
                     "config": {},
                     "next_step": "complete",
                     "output_field": "export_result",
                     "description": "Query llm_call_log, clean responses, stream into training_exports"
                   },
                   "complete": {
                     "action": "complete_workflow",
                     "config": {
                       "output_fields": ["export_result"]
                     }
                   }
                 }
               },
               "processing_mode": "task",
               "timeout_seconds": 1800
             }'::jsonb,
             true
         )
    ON CONFLICT (type, version) WHERE deleted_at IS NULL DO UPDATE
                                                                SET default_config   = EXCLUDED.default_config,
                                                                description      = EXCLUDED.description,
                                                                category         = EXCLUDED.category,
                                                                agent_category   = EXCLUDED.agent_category,
                                                                status           = EXCLUDED.status,
                                                                updated_at       = NOW();


-- ----------------------------------------------------------------------------
-- ORCHESTRATOR WRAPPER: training-data-export-orchestrator
-- ----------------------------------------------------------------------------
-- spawn_agent → call_agent → complete_workflow. Three steps, no real work.
-- Exists so the worker gets a dedicated pod instead of running inside one
-- of the main agent-chassis pods.
--
-- Input mapping uses the ? suffix for optional fields per doc 001 §"Map
-- fields individually". Required fields (agent_type, step_name) have no ?
-- so the call fails loudly if missing.
--
-- Uses target_role for call_agent lookup (not agent_type) per doc 001
-- §"How call_agent finds the spawned agent" — more reliable, decoupled
-- from the spawn step's output_field naming.
-- ----------------------------------------------------------------------------

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status,
    default_config, is_active
) VALUES (
             'training-data-export-orchestrator',
             'Training Data Export Orchestrator',
             'Thin wrapper that spawns training-data-exporter in a dedicated pod and waits for the export to complete. This is the agent triggered via Kafka for manual exports. Passes input_data through to the worker.',
             'orchestrator',
             'coordinator',
             'experimental',
             '{
               "workflow": {
                 "start_step": "spawn_exporter",
                 "steps": {
                   "spawn_exporter": {
                     "action": "spawn_agent",
                     "config": {
                       "role": "exporter",
                       "agent_type": "training-data-exporter"
                     },
                     "next_step": "call_exporter",
                     "output_field": "spawn_exporter",
                     "description": "Spawn dedicated pod for training-data-exporter"
                   },
                   "call_exporter": {
                     "action": "call_agent",
                     "config": {
                       "target_role": "exporter",
                       "input_mapping": {
                         "agent_type":           "input_data.agent_type",
                         "step_name":            "input_data.step_name",
                         "model_filter?":        "input_data.model_filter",
                         "include_fenced?":      "input_data.include_fenced",
                         "strict_json?":         "input_data.strict_json",
                         "min_response_length?": "input_data.min_response_length",
                         "max_rows?":            "input_data.max_rows",
                         "source_notes?":        "input_data.source_notes"
                       },
                       "timeout_seconds": 1200
                     },
                     "next_step": "complete",
                     "output_field": "export_result",
                     "description": "Invoke the exporter, wait for completion, forward result"
                   },
                   "complete": {
                     "action": "complete_workflow",
                     "config": {
                       "output_fields": ["export_result"]
                     }
                   }
                 }
               },
               "processing_mode": "orchestrator",
               "timeout_seconds": 1800
             }'::jsonb,
             true
         )
    ON CONFLICT (type, version) WHERE deleted_at IS NULL DO UPDATE
                                                                SET default_config   = EXCLUDED.default_config,
                                                                description      = EXCLUDED.description,
                                                                category         = EXCLUDED.category,
                                                                agent_category   = EXCLUDED.agent_category,
                                                                status           = EXCLUDED.status,
                                                                updated_at       = NOW();


-- ----------------------------------------------------------------------------
-- Verification
-- ----------------------------------------------------------------------------

SELECT type,
       agent_category,
       is_active,
       is_snapshot,
       jsonb_extract_path_text(default_config, 'processing_mode') as proc_mode,
       jsonb_extract_path_text(default_config, 'workflow', 'start_step') as start_step
FROM agent_definitions
WHERE type IN ('training-data-exporter', 'training-data-export-orchestrator')
  AND deleted_at IS NULL
ORDER BY type;
