-- ROLLBACK for 750. Restores copy-editor's on_approve to the shape it had
-- before the fan-out change.
--
-- NOTE what "before" was: `include_fields: ["copy_edit","page_target"]`, which
-- [MEASURED 2026-09-03] could never resolve — 42 field mentions across 21 items,
-- all history, ZERO present at spec top level. So this rollback restores a
-- BROKEN configuration on purpose. Use it only to get back to a known state
-- while something else is diagnosed, not as a repair.

BEGIN;

DO $$
DECLARE
  n int;
BEGIN
  SELECT count(*) INTO n
  FROM agent_definitions
  WHERE type = 'copy-editor' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '750 ROLLBACK: expected exactly 1 live copy-editor row, found %', n;
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,request_review,config,on_approve}',
      jsonb_build_object(
        'item_type',      'section_edit',
        'handler_agent',  'section-editor',
        'include_fields', jsonb_build_array('copy_edit', 'page_target')
      ),
      false
    ),
    updated_at = NOW()
WHERE type = 'copy-editor'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

DO $$
DECLARE
  cur jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'request_review'->'config'->'on_approve'
    INTO cur
  FROM agent_definitions
  WHERE type = 'copy-editor' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF cur ? 'fan_out_from' THEN
    RAISE EXCEPTION '750 ROLLBACK VERIFY: fan_out_from still present: %', cur;
  END IF;
  RAISE NOTICE '750 ROLLBACK: on_approve restored to the pre-fan-out shape';
END $$;

COMMIT;
