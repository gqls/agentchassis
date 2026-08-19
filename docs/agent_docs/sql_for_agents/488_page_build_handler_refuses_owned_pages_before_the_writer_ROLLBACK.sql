-- ROLLBACK for 488 — remove the early owned-page refusal from page-build-handler's
-- load_page_record step (both the opt-in key and the error routing).
--
-- Rows already written by the mechanism stay untouched, deliberately:
-- owned_page_review rows with refused_by='load_page_record' and wont_fix stamps
-- are a true record of refusals that did happen. After this rollback the refusal
-- happens at save_page_sections again (the pre-488 state) — the WASTE returns but
-- no protection is lost, because the save-path guard never moved.

BEGIN;

DO $$
DECLARE
    step jsonb;
    n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF n <> 1 THEN
        RAISE EXCEPTION '488 ROLLBACK: expected exactly 1 live page-build-handler row, found %', n;
    END IF;

    SELECT default_config #> '{workflow,steps,load_page_record}' INTO step
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF NOT (step->'config' ? 'refuse_owned_page') AND step->>'error_step' IS NULL THEN
        RAISE EXCEPTION '488 ROLLBACK: neither key is present — 488 is not applied here; nothing to roll back';
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,load_page_record}',
           ((default_config #> '{workflow,steps,load_page_record}') - 'error_step')
               #- '{config,refuse_owned_page}',
           false),
       updated_at = NOW()
 WHERE type = 'page-build-handler'
   AND is_active
   AND COALESCE(is_snapshot, false) = false
   AND deleted_at IS NULL;

DO $$
DECLARE
    step jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,load_page_record}' INTO step
      FROM agent_definitions
     WHERE type = 'page-build-handler' AND is_active
       AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF step->'config' ? 'refuse_owned_page' OR step ? 'error_step' THEN
        RAISE EXCEPTION '488 ROLLBACK VERIFY: a key survived the removal: %', step::text;
    END IF;
    IF step->'config'->>'site_id' <> 'site_record.site_id'
       OR step->>'next_step' <> 'check_page_found' THEN
        RAISE EXCEPTION '488 ROLLBACK VERIFY: pre-existing keys did not survive: %', step::text;
    END IF;

    RAISE NOTICE '488 ROLLBACK OK: load_page_record is back to the pre-488 shape; the refusal is at save_page_sections again';
END $$;

COMMIT;
