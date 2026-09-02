-- 715_ROLLBACK: remove cta_rank_anomaly from the completeness-discovery-agent
-- checks array — the clean undo for 715_enable_cta_rank_anomaly_check_HOLD.sql
-- (council round 1, debug_historian, corr 9faa2a23: verify and rollback are
-- separate files, so a bad apply has an undo path distinct from the forward
-- migration). Sidecar: the runner's SIDECAR_RE excludes _ROLLBACK files, and
-- council scope refuses them client-side — apply BY HAND only.
--
-- Removing the name is safe at any time relative to the image: a registered
-- check that is not in the array simply does not run. The reverse order (name
-- present, check unregistered) is the one that fails the whole discovery step,
-- which is why the forward file is _HOLD.
--
-- Open cta_rank_anomaly work items are NOT touched: they are
-- needs_human_review rows a person may still want to read; cancel by hand if
-- the rollback is because the check misfired.

BEGIN;

SELECT snapshot_agent('completeness-discovery-agent',
    '715_ROLLBACK: pre-removal');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_checks,config,checks}',
        (SELECT COALESCE(jsonb_agg(e), '[]'::jsonb)
         FROM jsonb_array_elements(default_config #> '{workflow,steps,run_checks,config,checks}') e
         WHERE e <> '"cta_rank_anomaly"'::jsonb)
    )
WHERE type = 'completeness-discovery-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'cta_rank_anomaly';

DO $$
DECLARE
    still boolean;
BEGIN
    SELECT (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'cta_rank_anomaly'
    INTO still
    FROM agent_definitions
    WHERE type = 'completeness-discovery-agent'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF COALESCE(still, false) THEN
        RAISE EXCEPTION '715_ROLLBACK: cta_rank_anomaly still present after removal';
    END IF;
END $$;

COMMIT;
