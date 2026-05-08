-- ============================================================================
-- 021_training_data_preparer_agent.sql
-- ============================================================================
-- training-data-preparer is the first worker called by model-trainer.
-- It does three things in one Go action invocation:
--   1. INSERTs a row into model_lifecycle.training_runs with status='pending',
--      capturing the hyperparameters and provenance.
--   2. Streams training_exports.rows for the given export_id as NDJSON
--      (replays the snapshot — option B from design discussion).
--   3. Uploads to s3://finetuning/datasets/{export_id}/training.jsonl via
--      the canonical storage.S3Client. Returns the s3:// URI.
--
-- The Go action prepare_training_data_action.go is written separately.
-- This SQL just declares the agent and its (trivial) workflow.
--
-- Inputs (CollectedData["input_data"]):
--   export_id          UUID      — required
--   hyperparameters    JSONB     — required
--   triggered_by       UUID      — optional
--   orchestration_id   UUID      — optional
--
-- Outputs (final CollectedData):
--   preparation_result.training_run_id   UUID
--   preparation_result.dataset_uri       TEXT  (s3://...)
-- ============================================================================

INSERT INTO agent_definitions (
    id,
    type,
    display_name,
    description,
    category,
    agent_category,
    status,
    domain_tags,
    default_config,
    is_active,
    capabilities,
    image_repository,
    image_tag,
    resources,
    topics,
    health_config
)
VALUES (
           gen_random_uuid(),
           'training-data-preparer',
           'Training Data Preparer',
           'Worker: streams training_exports.rows to S3 JSONL and inserts a model_lifecycle.training_runs row in pending state.',
           'specialist',
           'specialist',
           'experimental',
           '["training", "data-prep", "s3", "qlora"]'::jsonb,
           '{
             "workflow": {
               "start_step": "prepare_data",
               "steps": {
                 "prepare_data": {
                   "action": "prepare_training_data",
                   "config": {
                     "input_mapping": {
                       "export_id": "input_data.export_id",
                       "hyperparameters": "input_data.hyperparameters",
                       "triggered_by": "input_data.triggered_by",
                       "orchestration_id": "input_data.orchestration_id"
                     },
                     "s3_bucket": "finetuning",
                     "s3_key_template": "datasets/{export_id}/training.jsonl"
                   },
                   "output_field": "preparation_result",
                   "next_step": "complete",
                   "description": "INSERT training_runs row, stream rows to S3 JSONL, return {training_run_id, dataset_uri}"
                 },
                 "complete": {
                   "action": "complete_workflow",
                   "config": {
                     "output_fields": ["preparation_result"]
                   },
                   "description": "Return preparation_result to caller"
                 }
               }
             }
           }'::jsonb,
           true,
           '["postgres-read", "s3-write", "training-data"]'::jsonb,
           'docker.io/aqls/agent-chassis',
           'latest',
           '{"requests": {"cpu": "200m", "memory": "512Mi"}, "limits": {"cpu": "1", "memory": "2Gi"}}'::jsonb,
           '{"requests": "system.agent.training-data-preparer.requests", "responses": "system.agent.training-data-preparer.responses"}'::jsonb,
           '{"liveness_initial_delay": 30, "liveness_period": 30}'::jsonb
       );

-- Verify
SELECT type, category, agent_category, status, is_active
FROM agent_definitions
WHERE type = 'training-data-preparer'
  AND deleted_at IS NULL;
