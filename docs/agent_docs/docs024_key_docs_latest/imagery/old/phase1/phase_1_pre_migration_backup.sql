-- ============================================================================
-- Phase 1 imagery work — pre-migration backup
--
-- Backs up the design-discovery-agent row that phase_1_register_imagery_checks.sql
-- will modify. Run this BEFORE applying that migration.
--
-- Naming follows the per-agent scoped pattern from
-- 006_news_feed_pipeline_v2.md (agent_def_<short>_backup_<YYYYMMDD>) plus a
-- _pre_phase1_imagery suffix to disambiguate from any other 20260505 backups.
--
-- Discipline (per 009_model_infrastructure.md):
--   - No DROP TABLE IF EXISTS. If a name collides, that's the safety net.
-- ============================================================================

BEGIN;

CREATE TABLE agent_def_design_discovery_agent_backup_20260505_pre_phase1_imagery AS
SELECT * FROM agent_definitions
WHERE type = 'design-discovery-agent';

-- ----------------------------------------------------------------------------
-- Verification — should report live=1, backup=1
-- ----------------------------------------------------------------------------
SELECT 'design-discovery-agent' AS agent,
       (SELECT COUNT(*) FROM agent_definitions WHERE type = 'design-discovery-agent') AS live,
       (SELECT COUNT(*) FROM agent_def_design_discovery_agent_backup_20260505_pre_phase1_imagery) AS backup;

COMMIT;


-- ============================================================================
-- RESTORE — only if reverting Phase 1 SQL changes
-- ============================================================================
-- BEGIN;
-- UPDATE agent_definitions
-- SET default_config = (
--     SELECT default_config
--     FROM agent_def_design_discovery_agent_backup_20260505_pre_phase1_imagery
--     LIMIT 1
-- ),
--     updated_at = NOW()
-- WHERE type = 'design-discovery-agent';
-- COMMIT;
