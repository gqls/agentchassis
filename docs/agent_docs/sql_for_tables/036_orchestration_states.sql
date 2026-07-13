-- ============================================================================
-- Migration: Add site_id to orchestration_states
-- ============================================================================
-- Enables direct querying of orchestrations by site without JSONB search.
-- Set at creation time from input_data.site_id or site_record.site_id.
-- Nullable — not all orchestrations relate to a site (e.g. health checks).
-- ============================================================================

ALTER TABLE orchestration_states
    ADD COLUMN IF NOT EXISTS site_id UUID;

-- Index for direct site lookup (most common query pattern)
CREATE INDEX IF NOT EXISTS idx_orch_site_id
    ON orchestration_states(site_id) WHERE site_id IS NOT NULL;

-- Composite index for "active orchestrations for this site"
CREATE INDEX IF NOT EXISTS idx_orch_site_active
    ON orchestration_states(site_id, status)
    WHERE site_id IS NOT NULL AND status NOT IN ('COMPLETED', 'FAILED');

-- FK to sites table (optional — only if you want referential integrity)
-- Uncomment if desired:
-- ALTER TABLE orchestration_states
--     ADD CONSTRAINT fk_orch_site FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE SET NULL;

-- ============================================================================
-- Backfill from collected_data (covers historical orchestrations)
-- ============================================================================

-- Path 1: input_data.site_id (most common — dispatch loop, adoption trigger)
UPDATE orchestration_states
SET site_id = (collected_data->'input_data'->>'site_id')::uuid
WHERE site_id IS NULL
  AND collected_data->'input_data'->>'site_id' IS NOT NULL
  AND collected_data->'input_data'->>'site_id' != '';

-- Path 2: site_record.site_id (set after ensure_site_record step)
UPDATE orchestration_states
SET site_id = (collected_data->'site_record'->>'site_id')::uuid
WHERE site_id IS NULL
  AND collected_data->'site_record'->>'site_id' IS NOT NULL
  AND collected_data->'site_record'->>'site_id' != '';

-- Path 3: top-level site_id
UPDATE orchestration_states
SET site_id = (collected_data->>'site_id')::uuid
WHERE site_id IS NULL
  AND collected_data->>'site_id' IS NOT NULL
  AND collected_data->>'site_id' != '';

-- ============================================================================
-- Verification
-- ============================================================================

SELECT
    'total orchestrations' as metric,
    count(*) as value
FROM orchestration_states
UNION ALL
SELECT
    'with site_id',
    count(*)
FROM orchestration_states WHERE site_id IS NOT NULL
UNION ALL
SELECT
    'without site_id',
    count(*)
FROM orchestration_states WHERE site_id IS NULL;

-- Check gamedesign.uk specifically
SELECT orchestration_id, owner_agent_type, status, current_step,
       site_id, created_at
FROM orchestration_states
WHERE site_id = '15a6cb16-5a86-4541-a8e4-d7106239b6a4'
ORDER BY created_at DESC
    LIMIT 5;

