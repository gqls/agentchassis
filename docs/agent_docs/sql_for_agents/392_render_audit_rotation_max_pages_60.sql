-- 392 — raise the render-audit rotation's max_pages 25 -> 60 (bugs_open/242,
-- fix candidate 3: MITIGATION alongside the honesty fix, not the fix).
--
-- THE DEFECT (bugs_open/242): every site with more than 25 live pages gets a
-- truncated weekly sweep whose stored artefact reads exactly like a complete
-- one, and the unaudited tail is the SAME pages every week. Both rotation runs
-- to date were truncated (loancalculator.co.uk 25-of-27, robot-hands.com
-- 25-of-31) and neither said so.
--
-- THE REAL FIX is in code, same lane, and this migration does not depend on it:
-- the request now carries pages_total/truncated, the adapter echoes them into
-- the reply summary (the only artefact an awaited step keeps — RFC_012),
-- write_render_audit_findings stamps them into its durable result, and an
-- agent_error_log RENDER_AUDIT_TRUNCATED row lands before dispatch. Raising
-- the cap merely makes the current fleet actually-complete (largest site today
-- is 31 live pages); the honesty machinery is for the day a site outgrows 60.
--
-- Blast radius: render-audit-agent's audit step config only. A bigger cap costs
-- render-audit pod minutes (tens of sequential Chromium navigations), which is
-- the dedicated pod's whole purpose (owner ruling 2026-07-28, its own lane).
-- Config-only: live on the next rotation fire, no image roll, and safe against
-- ANY binary (old and new chassis both read max_pages the same way).

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,audit,config,max_pages}', '60'::jsonb),
    updated_at = NOW()
WHERE type = 'render-audit-agent'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

-- Verify with DO/RAISE — a SELECT verify cannot stop the COMMIT.
DO $$
DECLARE
  v_max int;
BEGIN
  SELECT (default_config->'workflow'->'steps'->'audit'->'config'->>'max_pages')::int
    INTO v_max
  FROM agent_definitions
  WHERE type = 'render-audit-agent'
    AND is_active
    AND COALESCE(is_snapshot, false) = false
    AND deleted_at IS NULL;

  IF v_max IS DISTINCT FROM 60 THEN
    RAISE EXCEPTION '392: render-audit-agent audit.config.max_pages is % (expected 60) — aborting', v_max;
  END IF;
END $$;

COMMIT;
