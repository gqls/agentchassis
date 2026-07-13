-- ============================================================================
-- Migration: snapshots use agent_definitions_backup instead of inline rows
-- ============================================================================
-- Supersedes 030_snapshot_as_column.sql (the JSONB-column variant). Uses the
-- existing agent_definitions_backup table for the snapshot/revert mechanism,
-- adding two columns for snapshot-specific metadata. Bulk-backup rows (rows
-- inserted by ad-hoc INSERT INTO ... SELECT) leave the new columns NULL;
-- snapshot_agent always sets snapshot_taken_at, so the two use-cases are
-- cleanly separable by `WHERE snapshot_taken_at IS NOT NULL`.
--
-- Why:
--   1. Audit found 8 Go query sites that read agent_definitions without
--      filtering is_active or is_snapshot. Two pick the wrong row when a
--      snapshot exists with version + 1000.
--   2. Patch UPDATEs filtered only on `type` matched both the live row and
--      the snapshot row, overwriting the snapshot's pre-patch state and
--      breaking revert.
--   3. Moving snapshots out of agent_definitions eliminates both problems
--      structurally — there are no snapshot rows in agent_definitions for
--      any query to mistakenly return.
--
-- What changes:
--   - agent_definitions_backup gains snapshot_taken_at and snapshot_reason
--   - snapshot_agent INSERTs into agent_definitions_backup
--   - revert_agent reads most-recent snapshot from agent_definitions_backup
--   - Existing snapshot rows in agent_definitions are deleted (they're
--     contaminated by the analyze_site patch and useless for revert)
--   - agent_snapshots view is rebuilt against the new shape
--
-- Apply with: \i 030_snapshot_to_backup_table.sql
-- ============================================================================

BEGIN;

-- ──────────────────────────────────────────────────────────────────────
-- 1. Pre-flight: confirm what we're about to migrate
-- ──────────────────────────────────────────────────────────────────────
SELECT
    'snapshot_rows_in_agent_definitions_before' AS metric,
    COUNT(*) AS value
FROM agent_definitions
WHERE is_snapshot = true
  AND deleted_at IS NULL;

SELECT
    'rows_in_agent_definitions_backup_before' AS metric,
    COUNT(*) AS value
FROM agent_definitions_backup;

-- ──────────────────────────────────────────────────────────────────────
-- 2. Add snapshot-specific columns to agent_definitions_backup
-- ──────────────────────────────────────────────────────────────────────
ALTER TABLE agent_definitions_backup
    ADD COLUMN IF NOT EXISTS snapshot_taken_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS snapshot_reason   TEXT,
    ADD COLUMN IF NOT EXISTS restored_at       TIMESTAMPTZ;

COMMENT ON COLUMN agent_definitions_backup.snapshot_taken_at IS
    'Set by snapshot_agent() to NOW(). NULL for rows inserted via ad-hoc bulk backup. Filter on IS NOT NULL to find snapshot_agent rows specifically.';

COMMENT ON COLUMN agent_definitions_backup.snapshot_reason IS
    'Optional human-readable reason the snapshot was taken (e.g. "before analyze_site vocabulary patch"). May be NULL.';

COMMENT ON COLUMN agent_definitions_backup.restored_at IS
    'Set by revert_agent() to NOW() when the snapshot is used to revert the live row. NULL means snapshot has not been used. Snapshots are preserved as audit trail, never deleted on revert.';

-- Partial index for revert_agent's "find latest unrestored snapshot" query
CREATE INDEX IF NOT EXISTS agent_definitions_backup_snapshot_lookup
    ON agent_definitions_backup (type, snapshot_taken_at DESC)
    WHERE snapshot_taken_at IS NOT NULL
    AND restored_at IS NULL;

-- ──────────────────────────────────────────────────────────────────────
-- 3. Delete the contaminated snapshot rows from agent_definitions
--
--    These rows have is_snapshot = true and were created by the old
--    snapshot_agent. Their default_config was overwritten by the recent
--    analyze_site patch (which used an over-broad WHERE clause), so they
--    can't restore anything useful. Safe to delete.
--
--    If you want to preserve them as audit (which has limited value given
--    contamination), copy them into agent_definitions_backup with
--    snapshot_taken_at = created_at first. Skipping that here.
-- ──────────────────────────────────────────────────────────────────────
DELETE FROM agent_definitions
WHERE is_snapshot = true
  AND deleted_at IS NULL;

SELECT
    'snapshot_rows_in_agent_definitions_after' AS metric,
    COUNT(*) AS value
FROM agent_definitions
WHERE is_snapshot = true;
-- Expect 0

-- ──────────────────────────────────────────────────────────────────────
-- 4. Rewrite snapshot_agent: INSERT into agent_definitions_backup
-- ──────────────────────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION snapshot_agent(
    p_agent_type TEXT,
    p_reason     TEXT DEFAULT NULL
)
RETURNS UUID AS $$
DECLARE
v_source_id UUID;
    v_current_version INT;
BEGIN
    -- Find current live definition
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

    -- Copy the live row into agent_definitions_backup with snapshot metadata.
    -- id, version, created_at, updated_at etc. are copied verbatim from the
    -- source row so the backup is a true point-in-time copy. The two new
    -- columns (snapshot_taken_at, snapshot_reason) carry the snapshot's
    -- own metadata.
INSERT INTO agent_definitions_backup (
    id, type, display_name, description, category, default_config,
    is_active, created_at, updated_at, deleted_at, capabilities,
    image_repository, image_tag, command, resources, topics,
    health_config, env_vars, version, previous_version_id,
    task_workflow, orchestrator_workflow, orchestration_workflow,
    delegation_preferences, agent_category, status, domain_tags,
    briefing_questionnaire, usage_count, is_snapshot,
    input_contract, output_contract, idle_timeout_seconds,
    snapshot_taken_at, snapshot_reason
)
SELECT
    id, type, display_name, description, category, default_config,
    is_active, created_at, updated_at, deleted_at, capabilities,
    image_repository, image_tag, command, resources, topics,
    health_config, env_vars, version, previous_version_id,
    task_workflow, orchestrator_workflow, orchestration_workflow,
    delegation_preferences, agent_category, status, domain_tags,
    briefing_questionnaire, usage_count, is_snapshot,
    input_contract, output_contract, idle_timeout_seconds,
    NOW(), p_reason
FROM agent_definitions
WHERE id = v_source_id;

RAISE NOTICE 'Snapshot captured: type=%, source_version=%, source_id=%, reason=%',
        p_agent_type, v_current_version, v_source_id,
        COALESCE(p_reason, '(none)');

RETURN v_source_id;
END;
$$ LANGUAGE plpgsql;

-- ──────────────────────────────────────────────────────────────────────
-- 5. Rewrite revert_agent: read most-recent unrestored snapshot, restore,
--    mark restored (audit trail preserved — never deletes snapshot rows).
-- ──────────────────────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION revert_agent(p_agent_type TEXT)
RETURNS VOID AS $$
DECLARE
v_snapshot_config JSONB;
    v_snapshot_taken  TIMESTAMPTZ;
    v_active_id       UUID;
BEGIN
    -- Find most recent UNRESTORED snapshot for this agent.
    -- Already-restored snapshots are kept as audit trail but skipped here.
SELECT default_config, snapshot_taken_at
INTO v_snapshot_config, v_snapshot_taken
FROM agent_definitions_backup
WHERE type = p_agent_type
  AND snapshot_taken_at IS NOT NULL
  AND restored_at IS NULL
ORDER BY snapshot_taken_at DESC
    LIMIT 1;

IF v_snapshot_config IS NULL THEN
        RAISE EXCEPTION 'No unrestored snapshot found for type % in agent_definitions_backup', p_agent_type;
END IF;

    -- Find current live row
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
    updated_at     = NOW()
WHERE id = v_active_id;

-- Mark snapshot as restored. Row is preserved for audit trail; future
-- revert_agent calls skip it via the `restored_at IS NULL` filter.
UPDATE agent_definitions_backup
SET restored_at = NOW()
WHERE type = p_agent_type
  AND snapshot_taken_at = v_snapshot_taken;

RAISE NOTICE 'Reverted: type=%, id=%, from snapshot taken at %',
        p_agent_type, v_active_id, v_snapshot_taken;
END;
$$ LANGUAGE plpgsql;

-- ──────────────────────────────────────────────────────────────────────
-- 6. Rebuild the agent_snapshots view against agent_definitions_backup.
--    Includes restored_at so operators can see which snapshots have
--    already been used. View returns ALL snapshots (used and unused).
-- ──────────────────────────────────────────────────────────────────────
DROP VIEW IF EXISTS agent_snapshots;

CREATE VIEW agent_snapshots AS
SELECT
    adb.type,
    adb.id                  AS source_id,
    adb.snapshot_taken_at   AS snapshot_taken,
    adb.restored_at,
    adb.snapshot_reason,
    adb.default_config->'workflow'->'steps' AS step_keys,
    s.key                   AS llm_step,
    s.value->'config'->'ai_service'->>'model'    AS snapshot_model,
        s.value->'config'->'ai_service'->>'provider' AS snapshot_provider
        FROM agent_definitions_backup adb
        LEFT JOIN LATERAL jsonb_each(adb.default_config->'workflow'->'steps') s(key, value)
        ON s.value->'config'->'ai_service' IS NOT NULL
        WHERE adb.snapshot_taken_at IS NOT NULL;

-- ──────────────────────────────────────────────────────────────────────
-- 7. Verify
-- ──────────────────────────────────────────────────────────────────────
SELECT 'remaining_snapshot_rows_in_agent_definitions' AS metric, COUNT(*) AS value
FROM agent_definitions
WHERE is_snapshot = true
UNION ALL
SELECT 'snapshot_rows_in_backup_table',
       COUNT(*)
FROM agent_definitions_backup
WHERE snapshot_taken_at IS NOT NULL
UNION ALL
SELECT 'bulk_backup_rows_in_backup_table',
       COUNT(*)
FROM agent_definitions_backup
WHERE snapshot_taken_at IS NULL;
-- Expect:
--   remaining_snapshot_rows_in_agent_definitions = 0
--   snapshot_rows_in_backup_table = 0 (no snapshots taken yet under new scheme)
--   bulk_backup_rows_in_backup_table = your existing ad-hoc backup count

COMMIT;

-- ──────────────────────────────────────────────────────────────────────
-- Smoke test (run outside the migration transaction)
-- ──────────────────────────────────────────────────────────────────────
-- Take a test snapshot and inspect:
--   SELECT snapshot_agent('site-adoption-agent', 'smoke test after migration');
--   SELECT type, snapshot_taken_at, snapshot_reason, restored_at,
--          LEFT(default_config->'workflow'->'steps'->'analyze_site'->'config'->>'prompt_template', 80) AS prompt_first_80
--   FROM agent_definitions_backup
--   WHERE snapshot_taken_at IS NOT NULL
--   ORDER BY snapshot_taken_at DESC LIMIT 1;
--
-- Revert and confirm the snapshot is now marked restored (NOT deleted):
--   SELECT revert_agent('site-adoption-agent');
--   SELECT type, snapshot_taken_at, restored_at, snapshot_reason
--   FROM agent_definitions_backup
--   WHERE type = 'site-adoption-agent' AND snapshot_taken_at IS NOT NULL
--   ORDER BY snapshot_taken_at DESC LIMIT 3;
--   -- Expect: latest row has restored_at populated; the live row is
--   -- unchanged because the snapshot config matched current config
--   -- (a no-op revert immediately after a snapshot).
--
-- Calling revert_agent again with no further snapshot should raise an
-- exception about no unrestored snapshot existing:
--   SELECT revert_agent('site-adoption-agent');
--   -- ERROR: No unrestored snapshot found for type site-adoption-agent...
--
-- ──────────────────────────────────────────────────────────────────────
-- Notes
-- ──────────────────────────────────────────────────────────────────────
-- • is_snapshot column on agent_definitions is intentionally NOT dropped.
--   Always false going forward; can drop in a follow-up migration once
--   confirmed nothing reads it. Lines 16080 and ~108516 of the chassis Go
--   code reference it; they'd need a small change to drop the column safely.
--
-- • Patches must still call snapshot_agent before applying. Caller contract
--   is unchanged (the new p_reason parameter is optional). The patch's
--   UPDATE on default_config no longer needs to filter is_snapshot — no
--   snapshot rows exist in agent_definitions for it to accidentally match.
--   Keeping the `AND is_active = true` filter is still good hygiene.
--
-- • Ad-hoc bulk backups (`INSERT INTO agent_definitions_backup SELECT * FROM
--   agent_definitions WHERE ...`) continue to work as before — they just
--   leave snapshot_taken_at and snapshot_reason NULL. revert_agent's
--   "find latest snapshot" query filters on snapshot_taken_at IS NOT NULL
--   so it never returns a bulk-backup row.
