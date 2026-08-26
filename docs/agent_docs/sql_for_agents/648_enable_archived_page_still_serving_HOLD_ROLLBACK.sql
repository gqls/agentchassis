-- ROLLBACK for 648 — remove archived_page_still_serving from
-- availability-discovery-agent's checks array.
--
-- WHEN YOU WOULD RUN THIS: the check is filing findings you do not trust (a
-- false-positive class its header's stated gaps did not anticipate — the most
-- likely is a WAF answering the real page with a 200 challenge body while
-- correctly 404ing the invented-URL control), or it is refusing so often that
-- the DISCOVERY_CHECK_ERROR rows are noise rather than signal, or the roll it
-- depends on was reverted so the binary no longer registers the name.
--
-- NOTE THE ASYMMETRY WITH APPLYING. Removing a name is always safe: an absent
-- check is simply not run. It is ADDING a name the binary does not know that
-- fails the whole discovery step and discards every earlier check's findings in
-- the same run. So this file needs no hold and no capability probe.
--
-- The Go check itself stays registered in the binary and is inert once its name
-- is out of the array — reverting the code is a separate decision and is not
-- what this file does.
--
-- ⚠ WHAT THIS DOES **NOT** UNDO: any archived_page_still_serving work items
-- already filed. They are flag-only records of a real observation and stay
-- open for a human. If you also mean to withdraw them, close them by hand with
-- a stated reason — do NOT leave them to the stale reaper, which would make the
-- withdrawal indistinguishable from a defect nobody looked at.
--
-- COMPANION EDIT: remove "archived_page_still_serving" from liveConfiguredChecks
-- in discovery_checks_registration_test.go in the same commit that runs this,
-- for the same reason the forward migration adds it — that fixture asserts the
-- live roster.

BEGIN;

SELECT snapshot_agent('availability-discovery-agent',
  '648_ROLLBACK_archived_page_still_serving: pre-revert');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (SELECT COALESCE(jsonb_agg(v), '[]'::jsonb)
            FROM jsonb_array_elements(
                   default_config #> '{workflow,steps,run_checks,config,checks}') AS v
           WHERE v <> '"archived_page_still_serving"'::jsonb)),
       updated_at = NOW()
 WHERE type = 'availability-discovery-agent' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'archived_page_still_serving';

-- Induced verification: the name must be gone, AND the array must not have been
-- emptied by a COALESCE mishap on the way. A checks array of [] would silently
-- disable availability discovery entirely — a worse outcome than the findings
-- this file is backing out.
DO $$
DECLARE still int; remaining int;
BEGIN
  SELECT count(*) INTO still
    FROM agent_definitions
   WHERE type = 'availability-discovery-agent' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'archived_page_still_serving';
  IF still <> 0 THEN
    RAISE EXCEPTION '648_ROLLBACK: archived_page_still_serving still present on % row(s)', still;
  END IF;

  SELECT jsonb_array_length(default_config #> '{workflow,steps,run_checks,config,checks}')
    INTO remaining
    FROM agent_definitions
   WHERE type = 'availability-discovery-agent' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF remaining <> 2 THEN
    RAISE EXCEPTION '648_ROLLBACK: checks array is % long, expected exactly 2 (site_unreachable + page_content_divergence) — the removal took something else with it', remaining;
  END IF;
END $$;

COMMIT;
