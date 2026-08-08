-- 340_unbuilt_link_dispatch_authoritative_page_id.sql
--
-- bugs_open/220: the build-dispatch-loop's call_handler maps page_name from
-- current_item.spec.page_name, which for unbuilt_internal_link is the page
-- CONTAINING the link — the handler rebuilds/redeploys the container, returns
-- success, and the never-deployed TARGET (the item's page_id column, where the
-- check deliberately files it) stays a live 404. Two config halves, matching
-- the Go halves that ride the next chassis roll:
--
--   1. build-dispatch-loop call_handler.input_mapping gains
--      "page_id?": "current_item.page_id" — forwards the work item's own
--      actionable-page column to every handler this loop dispatches.
--      PRECEDENT: site-work-orchestrator's call_handler already maps
--      "page_id?": "current_fix_item.page_id"; this loop was the outlier.
--      Optional key: a NULL column stays ABSENT (never ''), per the
--      bugs_open/154 rule.
--   2. page-build-handler load_page_record.config gains
--      "authoritative_page_id": "input_data.page_id" — the opt-in field
--      (owner ruling 2026-08-02, RFC_010 §2) that makes an explicitly
--      forwarded work-item id WIN over the name-first resolution that
--      otherwise always picks the container.
--
-- ORDERING: both halves are inert against the pre-roll binary — no live
-- handler dispatched by this loop reads input_data.page_id (measured
-- 2026-08-08: page-retraction and deduplicate-sections read that path but have
-- zero work items ever and are not dispatched here), and the old binary's
-- LoadPageRecordInputSpec does not declare authoritative_page_id so
-- ExtractActionInputs never reads the key (declared-field iteration;
-- UnknownConfigKeys is an offline audit, not a runtime rejection). Safe to
-- apply immediately; arms on the roll that carries the Go.
--
-- BLAST RADIUS, measured before writing (bug file § CONTRIB 2026-08-08, 116
-- lane): over all live work items, unbuilt_internal_link is the ONLY item type
-- whose page_id column names a different page from spec.page_name — every
-- other type either carries no column (key stays absent) or agrees with the
-- name (same page loads either way).
--
-- ROLLBACK: 340_..._ROLLBACK.sql — removes both keys.

\set ON_ERROR_STOP on

BEGIN;

-- Guard: refuse to run twice / against drift (targeted state, not blind set).
DO $pre$
DECLARE
    v_mapping jsonb;
    v_lpr     jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping}'
      INTO v_mapping
      FROM agent_definitions
     WHERE type = 'build-dispatch-loop'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_mapping IS NULL THEN
        RAISE EXCEPTION '340: no active build-dispatch-loop call_handler input_mapping found — nesting drifted, read the live row before applying';
    END IF;
    IF v_mapping ? 'page_id?' THEN
        RAISE EXCEPTION '340: input_mapping already carries page_id? (already applied?): %', v_mapping;
    END IF;
    IF v_mapping #>> '{page_name?}' IS DISTINCT FROM 'current_item.spec.page_name' THEN
        RAISE EXCEPTION '340: page_name? mapping is not the expected current_item.spec.page_name — live config drifted: %', v_mapping;
    END IF;

    SELECT default_config #> '{workflow,steps,load_page_record,config}'
      INTO v_lpr
      FROM agent_definitions
     WHERE type = 'page-build-handler'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_lpr IS NULL THEN
        RAISE EXCEPTION '340: no active page-build-handler load_page_record step found';
    END IF;
    IF v_lpr ? 'authoritative_page_id' THEN
        RAISE EXCEPTION '340: load_page_record already carries authoritative_page_id (already applied?): %', v_lpr;
    END IF;
END $pre$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping,page_id?}',
           '"current_item.page_id"'::jsonb
       ),
       updated_at = NOW()
 WHERE type = 'build-dispatch-loop'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,load_page_record,config,authoritative_page_id}',
           '"input_data.page_id"'::jsonb
       ),
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify with DO/RAISE — a SELECT-only verify block cannot stop the COMMIT.
DO $post$
DECLARE
    v_mapping jsonb;
    v_lpr     jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,process_item,config,sub_workflow,steps,call_handler,config,input_mapping}'
      INTO v_mapping
      FROM agent_definitions
     WHERE type = 'build-dispatch-loop'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_mapping #>> '{page_id?}' IS DISTINCT FROM 'current_item.page_id' THEN
        RAISE EXCEPTION '340 FAILED: page_id? mapping missing or wrong after update: %', v_mapping;
    END IF;

    SELECT default_config #> '{workflow,steps,load_page_record,config}'
      INTO v_lpr
      FROM agent_definitions
     WHERE type = 'page-build-handler'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_lpr #>> '{authoritative_page_id}' IS DISTINCT FROM 'input_data.page_id' THEN
        RAISE EXCEPTION '340 FAILED: authoritative_page_id missing or wrong after update: %', v_lpr;
    END IF;

    RAISE NOTICE '340 OK: dispatcher forwards current_item.page_id; page-build-handler load_page_record honours it';
END $post$;

COMMIT;
