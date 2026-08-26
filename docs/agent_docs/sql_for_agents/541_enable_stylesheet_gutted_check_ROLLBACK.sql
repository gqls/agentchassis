-- ROLLBACK for 541 — remove stylesheet_gutted from design-discovery-agent's
-- checks array.
--
-- When you would run this: the check is filing findings you do not trust (a false
-- positive class the header's stated gaps did not anticipate — a property defined
-- only by runtime JavaScript, or by a stylesheet on another host), or the roll it
-- depends on was reverted so the binary no longer registers the name.
--
-- NOTE THE ASYMMETRY WITH APPLYING. Removing a name is always safe: an absent
-- check is simply not run. It is ADDING a name the binary does not know that
-- fails the whole discovery step and discards earlier checks' findings
-- (discovery_checks.go:198-216). So this file needs no hold and no capability
-- probe.
--
-- The Go check itself stays registered in the binary and is inert once its name
-- is out of the array — reverting the code is a separate decision and is not what
-- this file does.
--
-- COMPANION EDIT: remove "stylesheet_gutted" from liveConfiguredChecks in
-- discovery_checks_registration_test.go in the same commit that runs this, for
-- the same reason the forward migration adds it — that fixture asserts the live
-- roster.

BEGIN;

SELECT snapshot_agent('design-discovery-agent',
  '541_ROLLBACK_stylesheet_gutted_check: pre-revert');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,run_checks,config,checks}',
         (SELECT COALESCE(jsonb_agg(v), '[]'::jsonb)
            FROM jsonb_array_elements(
                   default_config #> '{workflow,steps,run_checks,config,checks}') AS v
           WHERE v <> '"stylesheet_gutted"'::jsonb))
 WHERE type = 'design-discovery-agent' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'stylesheet_gutted';

-- Induced verification: the name must be gone, and the array must not have been
-- emptied by a COALESCE mishap on the way (a check array of [] would silently
-- disable design discovery entirely, which is a worse outcome than the finding
-- this file is backing out).
DO $$
DECLARE still int; remaining int;
BEGIN
  SELECT count(*) INTO still
    FROM agent_definitions
   WHERE type = 'design-discovery-agent' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND (default_config #> '{workflow,steps,run_checks,config,checks}') ? 'stylesheet_gutted';
  IF still <> 0 THEN
    RAISE EXCEPTION '541_ROLLBACK: stylesheet_gutted still present on % row(s)', still;
  END IF;

  SELECT jsonb_array_length(default_config #> '{workflow,steps,run_checks,config,checks}')
    INTO remaining
    FROM agent_definitions
   WHERE type = 'design-discovery-agent' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF remaining IS NULL OR remaining < 20 THEN
    RAISE EXCEPTION '541_ROLLBACK: checks array is % entries — expected the other 23 to survive; investigate before committing', remaining;
  END IF;
END $$;

COMMIT;
