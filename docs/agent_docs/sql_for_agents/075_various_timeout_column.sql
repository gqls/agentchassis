-- ============================================================================
-- Add idle_timeout_seconds to agent_definitions
--
-- Controls how long a spawned agent (K8s Job) waits with no messages
-- before exiting cleanly. 0 = no timeout (for Deployment agents).
--
-- When the pod exits, K8s marks the Job as Complete, and
-- TTLSecondsAfterFinished cleans it up.
-- ============================================================================

ALTER TABLE agent_definitions
    ADD COLUMN idle_timeout_seconds int NOT NULL DEFAULT 0;

-- All agent types that get spawned as Jobs should auto-exit.
-- 120s is generous — most inter-step gaps are under 30s.
-- The timer resets on every message (request or response), so a
-- multi-step workflow with sub-agent calls stays alive as long as
-- responses keep arriving.

UPDATE agent_definitions
SET idle_timeout_seconds = 120
WHERE type IN (
               'build-dispatch-loop',
               'page-build-handler',
               'page-content-writer',
               'content-gap-planner',
               'component-template-fixer',
               'webdesign-agent',
               'rerender-pages',
               'page-rerender',
               'asset-deployer',
               'color-variable-fixer',
               'nav-link-fixer',
               'research-agent',
               'improvement-loop',
               'site-component-linker',
               'quality-discovery',
               'design-discovery',
               'completeness-discovery',
               'design-audit',
               'content-audit',
               'ux-audit'
    ) AND deleted_at IS NULL;

-- The vet pipeline agents also get spawned as Jobs
UPDATE agent_definitions
SET idle_timeout_seconds = 120
WHERE type IN (
               'vet-batch-processor',
               'vet-practice-verifier',
               'vet-pipeline-orchestrator',
               'area-sweep-orchestrator',
               'area-sweep-discoverer'
    ) AND deleted_at IS NULL;

-- Verify — Deployment agents should have 0, Job agents should have 120
SELECT type, idle_timeout_seconds,
       CASE WHEN idle_timeout_seconds > 0 THEN 'job' ELSE 'deployment' END as lifecycle
FROM agent_definitions
WHERE deleted_at IS NULL AND is_active = true
ORDER BY idle_timeout_seconds DESC, type;