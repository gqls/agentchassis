-- ROLLBACK for 480 — remove the owned-page refusal opt-in from page-build-handler.
--
-- Effect of reverting: an ownership refusal at save_sections goes back to writing
-- `failed`, and so goes back to counting against the (item_type, handler_agent)
-- pair in the detected-item-promoter's 25% floor. That is the defect 480 exists to
-- fix, so revert only if the substitution is causing harm — the likeliest reason
-- being that some consumer treats `wont_fix` in a way this lane did not find.
--
-- What this does NOT do: rows already written `wont_fix` stay `wont_fix`. They are
-- a true record of a refusal that did happen, and rewriting history to `failed`
-- would inject exactly the false votes 480 removed. If you genuinely need them
-- counted again, they carry `result->'owned_page_refusal' = true` and can be found
-- with:
--     SELECT id, item_type, handler_agent, updated_at
--       FROM site_work_items
--      WHERE status = 'wont_fix' AND result ? 'owned_page_refusal';
--
-- The Go half stays in place and stays inert: with the key absent,
-- update_work_item_status never substitutes anything. No roll is needed to revert.

BEGIN;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,mark_item_failed,config}' INTO cfg
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '480 ROLLBACK: no live page-build-handler mark_item_failed step';
    END IF;

    IF NOT (cfg ? 'owned_page_refusal_status') THEN
        RAISE EXCEPTION '480 ROLLBACK: the key is already absent — 480 was never applied, or has already been rolled back';
    END IF;

    IF cfg->>'owned_page_refusal_status' <> 'wont_fix' THEN
        RAISE EXCEPTION '480 ROLLBACK: the key holds % , not the wont_fix 480 wrote — someone has changed it deliberately; decide before removing it', cfg->>'owned_page_refusal_status';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = default_config #- '{workflow,steps,mark_item_failed,config,owned_page_refusal_status}',
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,mark_item_failed,config}' INTO cfg
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg ? 'owned_page_refusal_status' THEN
        RAISE EXCEPTION '480 ROLLBACK VERIFY: the key is still present';
    END IF;

    -- The rest of the step must be untouched — a #- on the wrong path would
    -- otherwise look like a clean rollback.
    IF cfg->>'status' <> 'failed'
       OR cfg->>'work_item_id_field' <> 'input_data.work_item_id'
       OR (cfg->>'skip_if_missing')::boolean IS DISTINCT FROM true THEN
        RAISE EXCEPTION '480 ROLLBACK VERIFY: the step lost a pre-existing key: %', cfg::text;
    END IF;

    RAISE NOTICE '480 ROLLBACK OK: the opt-in is gone, the step is otherwise unchanged';
END $$;

COMMIT;
