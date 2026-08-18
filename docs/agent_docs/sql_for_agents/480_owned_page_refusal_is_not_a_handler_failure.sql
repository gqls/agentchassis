-- 480 — page-build-handler: an OWNED-PAGE REFUSAL records `wont_fix`, not `failed`
--       (OWNER DECISION 2026-08-18, decision 1: "do not switch the handler off for
--        this — write something other than `failed`")
--
-- This is the CONFIG half of a two-part change. The Go half is in the same commit:
--   platform/orchestration/actions/save_page_sections_action.go  (marker on the refusal error)
--   platform/orchestration/actions/v3_site_actions.go            (the opt-in field this sets)
--
-- ============================================================================
-- WHAT THE PROBLEM IS, in plain terms
-- ============================================================================
-- A page whose `pages.rebuild_policy` is 'owned' belongs to a tool or a widget.
-- The owned-page guard therefore refuses a generic section save over it, because
-- that save DELETEs and reinserts every page_component and would clobber the live
-- tool. Refusing is correct and is not in question.
--
-- What is wrong is what the refused WORK ITEM then records. It dies `failed` —
-- the same word used for a handler that tried the job and could not do it.
--
-- ============================================================================
-- THE RULE THAT MAKES THAT EXPENSIVE
-- ============================================================================
-- `scheduled_tasks.detected-item-promoter` (SCH-026, migration 444, corrected by
-- 454 and 465) will not dispatch an (item_type, handler_agent) pair whose lifetime
-- successes are under 25% of its terminal outcomes, once there are >= 5 of them.
--
-- [MEASURED 2026-08-18, read from the live pre_query, not from a doc] that floor is
-- computed over exactly two buckets:
--     c = count(*) FILTER (WHERE status IN ('complete','verified'))
--     f = count(*) FILTER (WHERE status = 'failed')
-- and it names NO other status.
--
-- ============================================================================
-- HOW THIS CASE MEASURES AGAINST IT
-- ============================================================================
-- So every ownership refusal is a vote in a competence measure it says nothing
-- about. `phantom_internal_link -> page-build-handler` is the live illustration:
-- 101 ok / 46 failed = 69% on GENERIC pages, and 0 ok / 14 failed on OWNED ones,
-- which is not a handler doing badly — it is a handler that was never permitted to
-- start. Blended, the pair reads 47%. It is above the floor today only because its
-- owned rows are a minority; ~134 non-terminal findings are queued behind this
-- refusal on owned pages right now. When a pair crosses the floor the promoter
-- stops dispatching that item type ENTIRELY — including the 69% of findings on
-- generic pages that were being repaired.
--
-- With `wont_fix` the pair reads NEVER TESTED HERE, which is the truth: `wont_fix`
-- is in neither bucket, so the refusal leaves both the numerator and the
-- denominator alone.
--
-- ============================================================================
-- WHY THIS TOUCHES ONE STEP AND NOT THE ACTION
-- ============================================================================
-- OWNER RULING 2026-08-02 §2: new authority on a shared seam ships as an opt-in
-- field whose UNSAFE side is the default, so the decision is visible to a reviewer
-- of the CALLER. `owned_page_refusal_status` is absent everywhere else and is
-- therefore inert everywhere else; this migration is the single deliberate opt-in.
--
-- THE CONSUMERS, ENUMERATED RATHER THAN ASSERTED (owner ruling 2026-07-29 §3).
-- The refusal reaches a work-item status write only from `save_page_sections`.
-- Every live agent with such a step, and where its error routes:
--     page-build-handler      save_sections -> mark_item_failed   (update_work_item_status: THIS one)
--     page-rerender           save_sections -> no error_step at all
--     tool-recreation-handler save_sections -> complete_error     (complete_workflow, writes no item status)
-- One applicable consumer, opted in here. The other two are unaffected by
-- construction, not by assumption.
--
-- ============================================================================
-- ORDER — AND WHY THIS ONE IS SAFE TO APPLY BEFORE THE ROLL
-- ============================================================================
-- The usual rule is image first, then config, because config naming an
-- unregistered ACTION fails at runtime. That does not apply here: this adds a
-- config KEY to an action that already exists. `update_work_item_status` reads its
-- config by explicit lookup and there is no strict/unknown-key validator anywhere
-- in the orchestration path (`update_work_item_status` has no RegisterActionInputSpec
-- at all), so a binary predating the Go half simply never looks the key up and
-- behaves exactly as today. It activates by itself at the next chassis roll.
--
-- ROLLBACK: 480_owned_page_refusal_is_not_a_handler_failure_ROLLBACK.sql removes
-- the key and asserts its absence. Rows already written `wont_fix` stay `wont_fix`;
-- that is deliberate, they are a true record of a refusal that did happen.

BEGIN;

-- GUARD: refuse unless the live step is the one this file was written against.
DO $$
DECLARE
    step jsonb;
    n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '480: expected exactly 1 live page-build-handler row, found %', n;
    END IF;

    SELECT default_config #> '{workflow,steps,mark_item_failed}' INTO step
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step IS NULL THEN
        RAISE EXCEPTION '480: page-build-handler has no mark_item_failed step — the workflow has been restructured since 2026-08-18; re-derive this migration';
    END IF;

    IF step->>'action' <> 'update_work_item_status' THEN
        RAISE EXCEPTION '480: mark_item_failed runs % , not update_work_item_status — the opt-in field would be read by nothing', step->>'action';
    END IF;

    IF step->'config'->>'status' <> 'failed' THEN
        RAISE EXCEPTION '480: mark_item_failed already writes % rather than failed — another lane has changed this step; re-read it before proceeding', step->'config'->>'status';
    END IF;

    IF step->'config' ? 'owned_page_refusal_status' THEN
        RAISE EXCEPTION '480: owned_page_refusal_status is ALREADY set (%) — another session has applied this or an equivalent; do not overwrite it', step->'config'->>'owned_page_refusal_status';
    END IF;

    -- The save step must actually route its error here, or the field sits on a
    -- step the refusal never reaches and this migration would look applied and
    -- do nothing.
    SELECT default_config #> '{workflow,steps,save_sections}' INTO step
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF COALESCE(step->'config'->>'error_step', step->>'error_step') <> 'mark_item_failed' THEN
        RAISE EXCEPTION '480: save_sections routes its error to %, not mark_item_failed — the opt-in belongs on whatever step it DOES route to', COALESCE(step->'config'->>'error_step', step->>'error_step', '(none)');
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,mark_item_failed,config,owned_page_refusal_status}',
           '"wont_fix"'::jsonb,
           true),
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

-- VERIFY — a DO block, not a SELECT. ON_ERROR_STOP does not fire on a non-empty
-- result set, so a verify block made of SELECTs cannot stop the COMMIT (see
-- LANDMINES / RFC_006). This one RAISEs.
DO $$
DECLARE
    cfg jsonb;
    described text;
BEGIN
    SELECT default_config #> '{workflow,steps,mark_item_failed,config}' INTO cfg
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg->>'owned_page_refusal_status' <> 'wont_fix' THEN
        RAISE EXCEPTION '480 VERIFY: owned_page_refusal_status is % after the update', COALESCE(cfg->>'owned_page_refusal_status', '(absent)');
    END IF;

    -- The existing keys must all survive. jsonb_set on a path is surgical, but
    -- asserting it is what turns "surgical" from a belief into a check.
    IF cfg->>'status' <> 'failed'
       OR cfg->>'work_item_id_field' <> 'input_data.work_item_id'
       OR (cfg->>'skip_if_missing')::boolean IS DISTINCT FROM true THEN
        RAISE EXCEPTION '480 VERIFY: the pre-existing config keys did not survive: %', cfg::text;
    END IF;

    -- NEGATIVE CONTROL, in the same transaction, or the assertion above passes
    -- identically on an UPDATE with no WHERE clause. No OTHER live agent may have
    -- acquired this key.
    SELECT string_agg(ad.type || '.' || s.key, ', ') INTO described
      FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
     WHERE ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
       AND s.value->'config' ? 'owned_page_refusal_status'
       AND NOT (ad.type = 'page-build-handler' AND s.key = 'mark_item_failed');

    IF described IS NOT NULL THEN
        RAISE EXCEPTION '480 VERIFY: the opt-in leaked to steps it was not meant for: %', described;
    END IF;

    RAISE NOTICE '480 OK: page-build-handler.mark_item_failed records wont_fix for an ownership refusal, failed for everything else; no other step carries the key';
END $$;

COMMIT;
