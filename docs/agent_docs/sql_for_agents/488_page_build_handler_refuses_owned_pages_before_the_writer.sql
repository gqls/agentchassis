-- 488 — page-build-handler: refuse an OWNED page at load_page_record, BEFORE the
--       LLM writer runs (bugs_open/301, fix candidate 1)
--
-- This is the CONFIG half of a two-part change. The Go half is in the same commit:
--   platform/orchestration/actions/load_page_record_action.go  (the refuse_owned_page opt-in)
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- page-build-handler's owned-page refusal lives only in save_page_sections — the
-- LAST step that touches the page. So a build item aimed at a
-- pages.rebuild_policy='owned' page spawns page-content-writer (an LLM call) and
-- the internal-link-resolver, runs both to completion, and is only then refused.
-- bugs_open/301 measured 39 such discarded chains in ~2.5 hours on one site; on
-- 2026-08-19 there were 146 open findings on owned pages routed at this handler,
-- each a guaranteed wasted chain under the old ordering.
--
-- ============================================================================
-- WHAT THIS DOES
-- ============================================================================
-- Two surgical edits to the live page-build-handler workflow:
--   1. load_page_record.config.refuse_owned_page = true — the Go opt-in: the
--      action now refuses a loaded page whose rebuild_policy is 'owned', filing
--      the same deduped owned_page_review row the save-path guard files
--      (refused_by='load_page_record') and erroring with the OWNED_PAGE_GUARD
--      marker at the FRONT of the message.
--   2. load_page_record.error_step = 'mark_item_failed' — routes that refusal
--      (and any genuine load error) to the step that already carries
--      owned_page_refusal_status='wont_fix' (migration 480), so the refusal
--      lands wont_fix with zero new vocabulary and the promoter's floor stays
--      blind to it. A genuine load error still lands failed — the
--      discrimination is the marker, matched in update_work_item_status.
--
-- The save-path guard STAYS: it is the backstop for every other caller and for
-- the early guard's fail-open window. Removing it would re-open bugs_closed/295.
--
-- THE CONSUMERS, ENUMERATED RATHER THAN ASSERTED (owner ruling 2026-07-29 §3).
-- Live agents carrying a load_page_record step, read from agent_definitions
-- 2026-08-19:
--     page-build-handler       — opted in HERE, the single deliberate opt-in
--     tool-recreation-handler  — the tool pipeline, the legitimate owner of
--                                owned pages: does NOT carry the key, and the
--                                negative control below asserts it stays that way
--
-- NAMED BEHAVIOUR CHANGE beyond the refusal: a genuine load_page_record error
-- (DB failure, malformed authoritative id) now routes to mark_item_failed — the
-- item dies visibly failed, attempt counted — instead of failing the workflow.
-- That aligns the step with every sibling in this workflow (plan_sections,
-- call_content_writer, save_sections, deploy_page all route there already).
--
-- ============================================================================
-- ORDER — SAFE TO APPLY BEFORE THE ROLL, deliberately
-- ============================================================================
-- load_page_record's input spec has CheckConfig without StrictConfig, so a binary
-- predating the Go half logs a deduped "unrecognised config key" warning and
-- ignores it (platform/validation/workflow.go:185-195 warns, never rejects). The
-- error_step is honoured by the existing coordinator. So: applied before the
-- roll, only the error-routing alignment activates; the refusal itself activates
-- at the next chassis roll. Applied after, the dormant code wakes. No _HOLD.
--
-- ROLLBACK: 488_..._ROLLBACK.sql removes both keys and asserts their absence.
-- owned_page_review rows and wont_fix stamps already written stay — they are a
-- true record of refusals that did happen.

BEGIN;

-- GUARD: refuse unless the live workflow is the one this file was written against.
DO $$
DECLARE
    step jsonb;
    mark jsonb;
    n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '488: expected exactly 1 live page-build-handler row, found %', n;
    END IF;

    SELECT default_config #> '{workflow,steps,load_page_record}' INTO step
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '488: page-build-handler has no load_page_record step — the workflow has been restructured since 2026-08-19; re-derive this migration';
    END IF;
    IF step->>'action' <> 'load_page_record' THEN
        RAISE EXCEPTION '488: the load_page_record step runs %, not the load_page_record action — the opt-in would be read by nothing', step->>'action';
    END IF;
    IF step->'config' ? 'refuse_owned_page' THEN
        RAISE EXCEPTION '488: refuse_owned_page is ALREADY set (%) — another session has applied this or an equivalent; do not overwrite it', step->'config'->>'refuse_owned_page';
    END IF;
    IF COALESCE(step->>'error_step', step->'config'->>'error_step') IS NOT NULL THEN
        RAISE EXCEPTION '488: load_page_record already routes errors to % — another lane has changed this step; re-read it before proceeding', COALESCE(step->>'error_step', step->'config'->>'error_step');
    END IF;
    IF step->>'next_step' <> 'check_page_found' THEN
        RAISE EXCEPTION '488: load_page_record next_step is %, not check_page_found — re-derive against the live workflow', step->>'next_step';
    END IF;

    -- The refusal's landing step must exist AND carry the Tier 1 opt-in, or the
    -- refusal would land `failed` and re-poison the promoter's floor — the exact
    -- state migration 480 exists to end.
    SELECT default_config #> '{workflow,steps,mark_item_failed}' INTO mark
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF mark IS NULL OR mark->>'action' <> 'update_work_item_status' THEN
        RAISE EXCEPTION '488: mark_item_failed missing or wrong action (%) — the error route has moved; re-derive', COALESCE(mark->>'action','(no step)');
    END IF;
    IF mark->'config'->>'owned_page_refusal_status' IS DISTINCT FROM 'wont_fix' THEN
        RAISE EXCEPTION '488: mark_item_failed does not carry owned_page_refusal_status=wont_fix (has %) — apply/verify migration 480 first, or an early refusal lands failed and poisons the promoter floor', COALESCE(mark->'config'->>'owned_page_refusal_status','(absent)');
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           jsonb_set(
               default_config,
               '{workflow,steps,load_page_record,config,refuse_owned_page}',
               'true'::jsonb,
               true),
           '{workflow,steps,load_page_record,error_step}',
           '"mark_item_failed"'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

-- VERIFY — a DO block that RAISEs, not SELECTs (a verify block of SELECTs cannot
-- stop the COMMIT; see LANDMINES / RFC_006).
DO $$
DECLARE
    step jsonb;
    described text;
BEGIN
    SELECT default_config #> '{workflow,steps,load_page_record}' INTO step
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF (step->'config'->>'refuse_owned_page')::boolean IS DISTINCT FROM true THEN
        RAISE EXCEPTION '488 VERIFY: refuse_owned_page is % after the update', COALESCE(step->'config'->>'refuse_owned_page', '(absent)');
    END IF;
    IF step->>'error_step' <> 'mark_item_failed' THEN
        RAISE EXCEPTION '488 VERIFY: error_step is % after the update', COALESCE(step->>'error_step', '(absent)');
    END IF;

    -- The existing keys must all survive. jsonb_set on a path is surgical, but
    -- asserting it is what turns "surgical" from a belief into a check.
    IF step->'config'->>'site_id'   <> 'site_record.site_id'
       OR step->'config'->>'page_name' <> 'input_data.spec.page_name'
       OR step->'config'->>'page_id'   <> 'input_data.spec.page_id'
       OR step->'config'->>'authoritative_page_id' <> 'input_data.page_id'
       OR step->>'next_step' <> 'check_page_found'
       OR step->>'output_field' <> 'page_record' THEN
        RAISE EXCEPTION '488 VERIFY: pre-existing load_page_record keys did not survive: %', step::text;
    END IF;

    -- NEGATIVE CONTROL, in the same transaction, or the assertions above pass
    -- identically on an UPDATE with no WHERE clause. No OTHER live agent or step
    -- may have acquired the key — tool-recreation-handler carrying it would put
    -- the tool pipeline in the position of refusing its own pages.
    SELECT string_agg(ad.type || '.' || s.key, ', ') INTO described
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
       AND s.value->'config' ? 'refuse_owned_page'
       AND NOT (ad.type = 'page-build-handler' AND s.key = 'load_page_record');

    IF described IS NOT NULL THEN
        RAISE EXCEPTION '488 VERIFY: the opt-in leaked to steps it was not meant for: %', described;
    END IF;

    RAISE NOTICE '488 OK: page-build-handler refuses an owned page at load_page_record (wont_fix via mark_item_failed); tool-recreation-handler untouched; save-path guard remains the backstop';
END $$;

COMMIT;
