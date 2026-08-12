-- 398 — the two 397 routers pass site_id to create_work_item explicitly
--
-- SUPERSEDES a gap in 397, which is already recorded in schema_migrations and
-- therefore must not be edited (the ledger's checksum is of the file as
-- applied — scripts/migration/README.md: "fix it, or supersede it with the next
-- number, rather than editing an already-recorded file").
--
-- WHAT WAS WRONG. `create_work_item` declares Required: ["site_id"]
-- (create_work_item_action.go:72). 397's two `create_work_item` steps set
-- twelve config keys and site_id was not among them, leaving it to
-- ExtractActionInputs' recursive search over collected_data to find
-- input_data.site_id. That may well work — but it is not what a single live
-- caller does. Measured 2026-08-12 across every active non-snapshot definition:
-- ELEVEN create_work_item steps across ten agents, and ALL ELEVEN set site_id
-- explicitly (seven as "input_data.site_id", four as "site_record.site_id").
-- Zero rely on the fallback. A required field resolved by an undeclared route
-- in a handler nobody can exercise yet is precisely the shape that fails
-- quietly months later, so it is being closed now while the agents are inert.
--
-- "input_data.site_id" is the correct half of that precedent for these two:
-- both are work-item handlers dispatched by build-dispatch-loop, whose
-- call_handler input_mapping supplies site_id at input_data.site_id. The
-- site_record.* spelling belongs to agents that load a site row first, which
-- neither of these does.
--
-- SURGICAL, not a regeneration: jsonb_set on the one key, so nothing else in
-- either workflow can move. The guard below asserts both the new key AND that
-- the step's other config survived — a regeneration-shaped mistake would pass a
-- check that only looked at site_id.
--
-- STILL INERT. This changes no assignment; both agents remain unreferenced by
-- any work item, per the owner's instruction to defer assignment until
-- bugs_open/248's rung 2 is cut.
--
-- Apply:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 398_….sql

BEGIN;

SELECT snapshot_agent('image-url-404-handler', '398_image_routers_set_site_id_explicitly.sql: pre-update');
SELECT snapshot_agent('image-source-unsatisfiable-handler', '398_image_routers_set_site_id_explicitly.sql: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,file_generation_request,config,site_id}',
           '"input_data.site_id"'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'image-url-404-handler'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,file_imagery_request,config,site_id}',
           '"input_data.site_id"'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'image-source-unsatisfiable-handler'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    cfg      jsonb;
    assigned integer;
BEGIN
    -- --- agent 1: the new key, AND the surroundings it must not have disturbed
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type = 'image-url-404-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg #>> '{workflow,steps,file_generation_request,config,site_id}' IS DISTINCT FROM 'input_data.site_id' THEN
        RAISE EXCEPTION '398: image-url-404-handler file_generation_request.site_id not set';
    END IF;
    IF cfg #>> '{workflow,steps,file_generation_request,config,item_type}' IS DISTINCT FROM 'needs_imagery' THEN
        RAISE EXCEPTION '398: image-url-404-handler file_generation_request lost item_type — jsonb_set overwrote more than the one key';
    END IF;
    IF cfg #> '{workflow,steps,file_generation_request,config,spec_paths}' IS NULL THEN
        RAISE EXCEPTION '398: image-url-404-handler file_generation_request lost spec_paths';
    END IF;
    IF cfg #>> '{workflow,steps,decide_route,config,else_step}' IS DISTINCT FROM 'escalate_deploy_path_mismatch' THEN
        RAISE EXCEPTION '398: image-url-404-handler lost its escalation branch';
    END IF;

    -- --- agent 2
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE type = 'image-source-unsatisfiable-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF cfg #>> '{workflow,steps,file_imagery_request,config,site_id}' IS DISTINCT FROM 'input_data.site_id' THEN
        RAISE EXCEPTION '398: image-source-unsatisfiable-handler file_imagery_request.site_id not set';
    END IF;
    IF cfg #>> '{workflow,steps,file_imagery_request,config,item_type}' IS DISTINCT FROM 'needs_imagery' THEN
        RAISE EXCEPTION '398: image-source-unsatisfiable-handler file_imagery_request lost item_type';
    END IF;
    IF cfg #> '{workflow,steps,file_imagery_request,config,spec_paths}' IS NULL THEN
        RAISE EXCEPTION '398: image-source-unsatisfiable-handler file_imagery_request lost spec_paths';
    END IF;
    IF cfg #>> '{workflow,steps,decide_route,config,else_step}' IS DISTINCT FROM 'escalate_unmappable_source' THEN
        RAISE EXCEPTION '398: image-source-unsatisfiable-handler lost its escalation branch';
    END IF;

    -- --- still inert
    SELECT count(*) INTO assigned FROM site_work_items
     WHERE handler_agent IN ('image-url-404-handler', 'image-source-unsatisfiable-handler');
    IF assigned <> 0 THEN
        RAISE EXCEPTION '398: % work item(s) now route to these handlers — assignment is still deferred until bugs_open/248 rung 2 is cut', assigned;
    END IF;

    RAISE NOTICE '398: site_id set explicitly on both routers; still INERT (0 assigned)';
END $$;

COMMIT;
