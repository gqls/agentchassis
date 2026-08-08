-- 337: directory-build-handler delegation must pass spec + current_page to page-build-handler
--
-- Second live-fire correction to migration 326 (bugs_open/206), same day as 336.
-- With 336's plain keys the delegated call passed the contract check and the
-- child built content — then failed at update_status:
--
--   step update_status failed: failed to execute action update_page_status:
--   could not determine page_id
--
-- Cause: page-build-handler's steps read the page name from THREE places, and
-- the delegation supplied only one. load_page_record reads input_data.page_name
-- (passed — why the build got this far); update_status reads
-- input_data.spec.page_name (its live config: page_name_field =
-- "input_data.spec.page_name") with a last-resort fallback to
-- current_page.name (v3_site_actions.go, UpdatePageStatusAction). The normal
-- dispatcher (build-dispatch-loop call_handler) supplies BOTH: "spec":
-- "current_item.spec" and "current_page": "current_item.spec". 326's delegation
-- supplied neither, so the child died one step before deploy_page —
-- content written and saved, nothing deployed, page still a 404.
--
-- Fix: mirror the dispatcher's own proven keys. The needs_page item's spec
-- carries page_name (verified on 715ec305: {"page_name": "directory-index",
-- "page_role": "entity-directory", ...}), so input_data.spec.page_name and
-- current_page (as the spec object) both resolve after this.
--
-- Scope: ONE jsonb path on the directory-build-handler row. Pipeline: build.
-- Config-only; live immediately on apply.

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_page_build_handler,config,input_mapping}',
        '{"site_id": "input_data.site_id", "domain": "input_data.domain", "page_name": "input_data.page_name", "spec": "input_data.spec", "current_page": "input_data.spec"}'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'directory-build-handler'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    mapping jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,call_page_build_handler,config,input_mapping}'
      INTO mapping
    FROM agent_definitions
    WHERE type = 'directory-build-handler'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF mapping IS NULL THEN
        RAISE EXCEPTION '337: no active directory-build-handler row (326 not applied?)';
    END IF;

    -- IS DISTINCT FROM (missing path -> NULL -> silent false under <>), per the
    -- landmine 528f545f6 fixed in 326's own guard.
    IF mapping #>> '{spec}' IS DISTINCT FROM 'input_data.spec' THEN
        RAISE EXCEPTION '337: spec mapping missing or wrong: %', mapping;
    END IF;
    IF mapping #>> '{current_page}' IS DISTINCT FROM 'input_data.spec' THEN
        RAISE EXCEPTION '337: current_page mapping missing or wrong: %', mapping;
    END IF;
    IF mapping #>> '{site_id}' IS DISTINCT FROM 'input_data.site_id'
       OR mapping #>> '{domain}' IS DISTINCT FROM 'input_data.domain'
       OR mapping #>> '{page_name}' IS DISTINCT FROM 'input_data.page_name' THEN
        RAISE EXCEPTION '337: a 336-era plain key was lost in the rewrite: %', mapping;
    END IF;

    RAISE NOTICE '337 OK: delegation passes spec + current_page alongside the plain keys';
END $$;

COMMIT;
