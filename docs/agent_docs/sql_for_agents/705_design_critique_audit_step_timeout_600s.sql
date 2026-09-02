-- 705_design_critique_audit_step_timeout_600s.sql — the audit sweep outgrew the 180s default.
--
-- WHY [MEASURED 2026-09-02]: the first after-leg critique run (work item e543ba1f,
-- orchestration 6cc3cd38) FAILED with __step_error {"message":"Request timed out
-- (code: TIMEOUT)","failed_step":"audit"} — while the render-audit adapter's real
-- response (contrast findings, 8 of 40 URLs, captures taken) arrived at
-- 17:26:24.57Z, ~182s after the orchestration was created at 17:23:22.06Z and
-- ~2s AFTER the default await timeout fired. The vision step never ran; the work
-- item still read 'complete' (the spawn-record trap, bugs 287).
--
-- MECHANISM: seed 645's audit step carries no timeout_seconds, so the awaited
-- request gets DefaultRequestTimeout = 180s (datahelpers/timeout_helpers.go:18).
-- The chain that honours this key end-to-end: ConvertStepTimeout
-- (timeout_helpers.go:23, reads config.timeout_seconds into step.Timeout, called
-- in executeStep before routing) -> getTimeout(step) at both awaited-request
-- builders (coordinator.go:2482, 2617). The 08-26 run passed because the sweep
-- was quicker before the hero batch made every page image-heavy; today it is a
-- ~2s coin-flip against the default.
--
-- WHY 600: measured sweep ~182s for max_pages 8 at 2 viewports; 600 is ~3.3x
-- headroom without approaching the 60-min workflow maxAge. The stuck-orchestration
-- takeover (5 min) applies only to StatusExecutingStep, not awaiting-responses
-- (approval steps await 24h by default on the same machinery).
--
-- ROLLBACK: remove the key — default reasserts itself:
--   UPDATE agent_definitions SET default_config =
--     default_config #- '{workflow,steps,audit,config,timeout_seconds}' WHERE ...;
-- or restore the snapshot taken below.

SELECT snapshot_agent('design-critique-agent',
                      '705_design_critique_audit_step_timeout_600s.sql: pre-update');

BEGIN;

DO $$
DECLARE n_defs integer; has_step_timeout boolean;
BEGIN
  SELECT count(*) INTO n_defs
  FROM agent_definitions
  WHERE type='design-critique-agent' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n_defs <> 1 THEN
    RAISE EXCEPTION 'MIGRATION 705: expected exactly 1 live design-critique-agent definition, found %', n_defs;
  END IF;

  SELECT (default_config->'workflow'->'steps'->'audit' ? 'timeout') INTO has_step_timeout
  FROM agent_definitions
  WHERE type='design-critique-agent' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF has_step_timeout THEN
    RAISE EXCEPTION 'MIGRATION 705: the audit step carries a workflow-level timeout key, which OUTRANKS config.timeout_seconds (ConvertStepTimeout respects step.Timeout first) — this write would be inert. Resolve that key instead.';
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,audit,config,timeout_seconds}',
        to_jsonb(600))
WHERE type='design-critique-agent' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE ts integer; has_step_timeout boolean;
BEGIN
  SELECT (default_config->'workflow'->'steps'->'audit'->'config'->>'timeout_seconds')::integer,
         (default_config->'workflow'->'steps'->'audit' ? 'timeout')
    INTO ts, has_step_timeout
  FROM agent_definitions
  WHERE type='design-critique-agent' AND is_active
    AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF ts IS DISTINCT FROM 600 OR has_step_timeout THEN
    RAISE EXCEPTION 'MIGRATION 705: audit step timeout_seconds is % (want 600) / outranking timeout key present: %', ts, has_step_timeout;
  END IF;
  RAISE NOTICE 'migration 705 OK: audit step awaits up to 600s (default was 180s)';
END $$;

COMMIT;
