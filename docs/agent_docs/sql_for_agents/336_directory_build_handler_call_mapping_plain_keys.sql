-- 336: directory-build-handler call_page_build_handler input_mapping — plain child-field keys
--
-- Corrects migration 326 (bugs_open/206), found in LIVE FIRE 2026-08-08: the
-- first real dispatch of the new handler (work item 715ec305-...,
-- needs_page:directory-index, revived and re-routed by the improvement loop's
-- own incomplete_page_group + refreshOpenWorkItem machinery) ran ensure_layout
-- successfully (site_plan_sections now carries directory-index: hero,
-- directory-listing) and then failed at call_page_build_handler:
--
--   contract violation for agent 'page-build-handler': missing required
--   fields: [site_id domain]
--   Provided fields: [input_data.domain input_data.page_name input_data.site_id]
--
-- Cause: 326 wrote input_mapping KEYS carrying the `input_data.` prefix. In a
-- call_agent input_mapping the KEY is the CHILD's field name; the VALUE is the
-- dot-path into the parent's data. 326 copied the value syntax onto both
-- sides. Working precedent for the plain-key shape: build-dispatch-loop's own
-- call_handler mapping ("site_id": "current_item.site_id", ...) and the
-- improvement-loop trigger's mapping ("site_id": "input_data.site_id") — the
-- very run that surfaced this bug used it successfully.
--
-- Scope: ONE jsonb path on the directory-build-handler row. Pipeline: build
-- (page-build dispatch). No Go change; live immediately on apply.

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_page_build_handler,config,input_mapping}',
        '{"site_id": "input_data.site_id", "domain": "input_data.domain", "page_name": "input_data.page_name"}'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'directory-build-handler'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
    mapping jsonb;
BEGIN
    SELECT default_config INTO cfg FROM agent_definitions
    WHERE type = 'directory-build-handler'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '336: no active directory-build-handler row (326 not applied?)';
    END IF;

    mapping := cfg #> '{workflow,steps,call_page_build_handler,config,input_mapping}';

    -- IS DISTINCT FROM, not <>/!= : a missing path returns NULL and a NULL
    -- comparison in IF is silently false (the exact landmine 528f545f6 fixed
    -- in 326's own guard).
    IF mapping #>> '{site_id}' IS DISTINCT FROM 'input_data.site_id' THEN
        RAISE EXCEPTION '336: plain-key site_id mapping missing or wrong: %', mapping;
    END IF;
    IF mapping #>> '{domain}' IS DISTINCT FROM 'input_data.domain' THEN
        RAISE EXCEPTION '336: plain-key domain mapping missing or wrong: %', mapping;
    END IF;
    IF mapping #>> '{page_name}' IS DISTINCT FROM 'input_data.page_name' THEN
        RAISE EXCEPTION '336: plain-key page_name mapping missing or wrong: %', mapping;
    END IF;
    IF mapping ? 'input_data.site_id' OR mapping ? 'input_data.domain' OR mapping ? 'input_data.page_name' THEN
        RAISE EXCEPTION '336: prefixed keys still present: %', mapping;
    END IF;

    RAISE NOTICE '336 OK: call_page_build_handler input_mapping uses plain child-field keys';
END $$;

COMMIT;
