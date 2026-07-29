-- Enable the orphan_element_refs discovery check on completeness-discovery-agent.
--
-- TIMED DELIBERATELY. The estate is at ZERO findings right now — all nine live
-- pages that carried this defect were repaired at source today and re-measured
-- with the real Go function (0 of 98). Turning a new detector on while its
-- population is empty means the first thing it ever reports is NEW breakage,
-- not a backlog, so a fire is unambiguous. That also satisfies the rule this
-- workstream learned the hard way: a check whose fire rate is unknown is
-- indistinguishable from a dead check, so inspect every early fire.
--
-- Requires the image carrying check_orphan_element_refs.go to be live first —
-- v1.0.1202, pod-verified on both replicas before this ran. A checks array
-- naming a check the binary does not register is a silent no-op.
\set ON_ERROR_STOP on
BEGIN;

DO $pre$
DECLARE v int;
BEGIN
    SELECT count(*) INTO v FROM agent_definitions
     WHERE type='completeness-discovery-agent' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
       AND default_config::text LIKE '%dead_controls%';
    IF v <> 1 THEN RAISE EXCEPTION 'expected exactly 1 live completeness-discovery-agent carrying a checks array, found %', v; END IF;
    SELECT count(*) INTO v FROM agent_definitions
     WHERE type='completeness-discovery-agent' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
       AND default_config::text LIKE '%orphan_element_refs%';
    IF v <> 0 THEN RAISE EXCEPTION 'orphan_element_refs is already enabled — another session got here first; read before re-adding'; END IF;
END $pre$;

UPDATE agent_definitions
   SET default_config = replace(default_config::text, '"dead_controls"', '"dead_controls", "orphan_element_refs"')::jsonb,
       updated_at = NOW()
 WHERE type='completeness-discovery-agent' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $verify$
DECLARE v int; v_dead int;
BEGIN
    SELECT count(*) INTO v FROM agent_definitions
     WHERE type='completeness-discovery-agent' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
       AND default_config::text LIKE '%orphan_element_refs%';
    IF v <> 1 THEN RAISE EXCEPTION 'check not enabled (% rows carry it)', v; END IF;

    -- The replace must have ADDED, not REPLACED: dead_controls must survive.
    SELECT count(*) INTO v_dead FROM agent_definitions
     WHERE type='completeness-discovery-agent' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
       AND default_config::text LIKE '%dead_controls%';
    IF v_dead <> 1 THEN RAISE EXCEPTION 'dead_controls was lost by the edit'; END IF;

    RAISE NOTICE 'orphan_element_refs enabled on completeness-discovery-agent; dead_controls intact';
END $verify$;

COMMIT;
