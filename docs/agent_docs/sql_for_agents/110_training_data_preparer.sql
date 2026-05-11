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


---
-- improve above

-- ============================================================================
-- 022_training_agents_corrections.sql
-- ============================================================================
-- Applies corrections from the development-guide compliance assessment:
--   1. agent_category set on model-trainer (was added retroactively before; this
--      makes it explicit so a fresh re-apply lands cleanly).
--   2. env_vars on training-data-preparer so the existing storage client
--      uses the finetuning bucket — no storage interface change required.
--      The chassis storage client reads IMAGE_BUCKET as a fallback (see
--      platform/storage/s3.go::NewS3Client). Spawned Job pods get this env
--      var from agent_definitions.env_vars at spawn time.
--   3. domain_tags consistency for both rows.
--
-- Step Zero check (per dev guide §0):
--   Existing agents searched: train, gpu, provision, ssh, vm, instance
--     - ch-detail-fetcher: external API call (Companies House) — different
--       enough (no GPU/binary transfer/job lifecycle); no reuse.
--     - med-url-mapper: external API (Firecrawl) — same conclusion.
--     - vet-practice-verifier: external HTTP lookup; no reuse.
--   Existing actions searched: prepare_training, run_training, provision,
--     ssh_exec, gpu, model_artefact, model_evaluation
--     - no precedents.
--   Decision: new agents necessary; no reuse.
-- ============================================================================

-- ── Correction 1: ensure model-trainer has agent_category set ────────────────
UPDATE agent_definitions
SET agent_category = 'coordinator',
    domain_tags = '["training", "orchestrator", "qlora", "thunder-compute"]'::jsonb,
    updated_at = NOW()
WHERE type = 'model-trainer'
  AND deleted_at IS NULL
  AND (agent_category IS NULL OR agent_category != 'coordinator');


-- ── Correction 2: training-data-preparer gets IMAGE_BUCKET=finetuning ────────
-- Spawned Job pods read this at startup; the chassis storage client falls
-- back to os.Getenv("IMAGE_BUCKET") when cfg.Bucket is empty (see s3.go).
-- This routes the worker's Upload() calls to the finetuning bucket without
-- modifying the storage interface or constructing a second S3Client.
UPDATE agent_definitions
SET env_vars = '[
        {"name": "IMAGE_BUCKET", "value": "finetuning"}
    ]'::jsonb,
    updated_at = NOW()
WHERE type = 'training-data-preparer'
  AND deleted_at IS NULL;


-- ── Verify ───────────────────────────────────────────────────────────────────
SELECT type, category, agent_category, status,
       env_vars,
       domain_tags
FROM agent_definitions
WHERE type IN ('model-trainer', 'training-data-preparer')
  AND deleted_at IS NULL
ORDER BY type;


---
-- updaging s3 paths etc
-- ============================================================================
-- 024_training_data_preparer_bucket_layout.sql
-- ============================================================================
-- Updates training-data-preparer's prepare_data step config to:
--   - use a generic top-level bucket: personae-model-training
--   - use finetuning/ as the first-level prefix inside it
-- Resulting URI:
--   s3://personae-model-training/finetuning/datasets/<export_id>/training.jsonl
--
-- This keeps the bucket reusable for adjacent model-training artefacts
-- (eval sets, base-model snapshots, intermediate checkpoints) under their
-- own first-level prefix paths, rather than spinning up a bucket per concern.
--
-- No Go change required: prepare_training_data action v3 already reads
-- `bucket` and `s3_key_template` from step config (with "finetuning" as the
-- default when bucket is unset).
--
-- PRE-REQ: bucket `personae-model-training` must exist on B2 with the
-- application key having read+write on it.
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,prepare_data,config,bucket}',
                '"personae-model-training"'::jsonb
        ),
        '{workflow,steps,prepare_data,config,s3_key_template}',
        '"finetuning/datasets/{export_id}/training.jsonl"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'training-data-preparer'
  AND deleted_at IS NULL;

-- Verify the bucket and key template are now in step config
SELECT type,
       default_config->'workflow'->'steps'->'prepare_data'->'config'->>'bucket' AS bucket,
    default_config->'workflow'->'steps'->'prepare_data'->'config'->>'s3_key_template' AS key_template
FROM agent_definitions
WHERE type = 'training-data-preparer'
  AND deleted_at IS NULL;

-- The IMAGE_BUCKET env_var on this agent (set in 022_training_agents_corrections.sql)
-- becomes a no-op: the action takes bucket from step config first, ignoring env.
-- Leaving it set is harmless. Removing it for cleanliness:
UPDATE agent_definitions
SET env_vars = '[]'::jsonb,
    updated_at = NOW()
WHERE type = 'training-data-preparer'
  AND deleted_at IS NULL;

SELECT type, env_vars
FROM agent_definitions
WHERE type = 'training-data-preparer'
  AND deleted_at IS NULL;