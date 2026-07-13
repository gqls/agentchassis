-- ============================================================================
-- Migration 083: Model Swap and Rollback Infrastructure
-- ============================================================================
-- Provides safe model swapping for agent definitions via:
--   1. snapshot_agent() — saves current config as is_snapshot=true row
--   2. swap_agent_model() — snapshots, then updates ai_service in a step
--   3. revert_agent() — restores from latest snapshot
--
-- These work alongside the manual backup table (agent_definitions_backup_*)
-- but are per-agent rather than full-table. The manual backup remains the
-- nuclear option.
--
-- Usage:
--   SELECT snapshot_agent('page-content-writer');
--   SELECT swap_agent_model('page-content-writer', 'generate_content',
--          '{"provider":"ollama","model":"llama3.3:70b","api_url":"http://ollama-gpu...:11434"}'::jsonb);
--   -- test, then if bad:
--   SELECT revert_agent('page-content-writer');
-- ============================================================================


-- ── 1. snapshot_agent ──
-- Takes a snapshot of the current active definition for a given agent type.
-- Returns the snapshot row's ID.

CREATE OR REPLACE FUNCTION snapshot_agent(p_agent_type TEXT)
RETURNS UUID AS $$
DECLARE
v_source_id UUID;
    v_snapshot_id UUID;
    v_current_version INT;
BEGIN
    -- Find current active definition
SELECT id, version INTO v_source_id, v_current_version
FROM agent_definitions
WHERE type = p_agent_type
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false)
ORDER BY version DESC
    LIMIT 1;

IF v_source_id IS NULL THEN
        RAISE EXCEPTION 'No active definition found for type %', p_agent_type;
END IF;

    -- Insert snapshot copy
INSERT INTO agent_definitions (
    type, display_name, description, category, default_config,
    is_active, capabilities, image_repository, image_tag,
    resources, topics, health_config, env_vars,
    version, previous_version_id, task_workflow,
    orchestrator_workflow, orchestration_workflow,
    delegation_preferences, agent_category, status,
    domain_tags, briefing_questionnaire, input_contract,
    output_contract, idle_timeout_seconds, is_snapshot
)
SELECT
    type, display_name, description, category, default_config,
    false,  -- snapshots are NOT active
    capabilities, image_repository, image_tag,
    resources, topics, health_config, env_vars,
    version + 1000,  -- offset to avoid unique constraint conflicts
    v_source_id,     -- points back to source
    task_workflow, orchestrator_workflow, orchestration_workflow,
    delegation_preferences, agent_category, 'active',
    domain_tags, briefing_questionnaire, input_contract,
    output_contract, idle_timeout_seconds, true  -- mark as snapshot
FROM agent_definitions
WHERE id = v_source_id
    RETURNING id INTO v_snapshot_id;

RAISE NOTICE 'Snapshot created: type=%, snapshot_id=%, source_version=%',
        p_agent_type, v_snapshot_id, v_current_version;

RETURN v_snapshot_id;
END;
$$ LANGUAGE plpgsql;


-- ── 2. swap_agent_model ──
-- Snapshots current config, then updates the ai_service block in a given step.
-- Returns the snapshot ID (for revert if needed).

CREATE OR REPLACE FUNCTION swap_agent_model(
    p_agent_type TEXT,
    p_step_name TEXT,
    p_new_ai_service JSONB
)
RETURNS UUID AS $$
DECLARE
v_snapshot_id UUID;
    v_current_ai JSONB;
BEGIN
    -- Verify the step exists and has ai_service
SELECT default_config->'workflow'->'steps'->p_step_name->'config'->'ai_service'
INTO v_current_ai
FROM agent_definitions
WHERE type = p_agent_type
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

IF v_current_ai IS NULL THEN
        RAISE EXCEPTION 'Step % not found or has no ai_service in agent %', p_step_name, p_agent_type;
END IF;

    -- Take snapshot first
    v_snapshot_id := snapshot_agent(p_agent_type);

    -- Apply the model swap
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        ARRAY['workflow', 'steps', p_step_name, 'config', 'ai_service'],
        p_new_ai_service
                     ),
    updated_at = NOW()
WHERE type = p_agent_type
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

RAISE NOTICE 'Model swapped: type=%, step=%, old=%, new=%, snapshot=%',
        p_agent_type, p_step_name,
        v_current_ai->>'model', p_new_ai_service->>'model',
        v_snapshot_id;

RETURN v_snapshot_id;
END;
$$ LANGUAGE plpgsql;


-- ── 3. revert_agent ──
-- Restores the active definition from the most recent snapshot.
-- Deletes the snapshot row after restoring.

CREATE OR REPLACE FUNCTION revert_agent(p_agent_type TEXT)
RETURNS VOID AS $$
DECLARE
v_snapshot_id UUID;
    v_snapshot_config JSONB;
    v_active_id UUID;
BEGIN
    -- Find most recent snapshot
SELECT id, default_config INTO v_snapshot_id, v_snapshot_config
FROM agent_definitions
WHERE type = p_agent_type
  AND is_snapshot = true
  AND deleted_at IS NULL
ORDER BY created_at DESC
    LIMIT 1;

IF v_snapshot_id IS NULL THEN
        RAISE EXCEPTION 'No snapshot found for type %', p_agent_type;
END IF;

    -- Find current active
SELECT id INTO v_active_id
FROM agent_definitions
WHERE type = p_agent_type
  AND is_active = true
  AND deleted_at IS NULL
  AND (is_snapshot IS NULL OR is_snapshot = false);

IF v_active_id IS NULL THEN
        RAISE EXCEPTION 'No active definition found for type %', p_agent_type;
END IF;

    -- Restore config from snapshot
UPDATE agent_definitions
SET default_config = v_snapshot_config,
    updated_at = NOW()
WHERE id = v_active_id;

-- Delete the used snapshot
DELETE FROM agent_definitions WHERE id = v_snapshot_id;

RAISE NOTICE 'Reverted: type=%, from snapshot=%', p_agent_type, v_snapshot_id;
END;
$$ LANGUAGE plpgsql;


-- ── 4. View: list all snapshots ──

CREATE OR REPLACE VIEW agent_snapshots AS
SELECT
    ad.type,
    ad.id as snapshot_id,
    ad.previous_version_id as source_id,
    ad.created_at as snapshot_taken,
    ad.default_config->'workflow'->'steps' as step_keys,
    s.key as llm_step,
    s.value->'config'->'ai_service'->>'model' as snapshot_model,
    s.value->'config'->'ai_service'->>'provider' as snapshot_provider
FROM agent_definitions ad
    LEFT JOIN LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s(key, value)
ON s.value->'config'->'ai_service' IS NOT NULL
WHERE ad.is_snapshot = true
  AND ad.deleted_at IS NULL
ORDER BY ad.type, ad.created_at DESC;


-- ── Verification ──

SELECT proname, pronargs
FROM pg_proc
WHERE proname IN ('snapshot_agent', 'swap_agent_model', 'revert_agent')
ORDER BY proname;

