-- ============================================================================
-- 370_retire_update_page_status_notes_and_validation_issues_fields_ROLLBACK.sql
--
-- Restores the two (dead) keys to content-reviewer.mark_page_needs_attention.
--
-- ⚠ Do NOT apply after the Go commit declaring these keys removed has rolled:
-- the binary rejects the whole content-reviewer workflow at validation on
-- every message while the keys are present. Roll back the code first.
-- ============================================================================

BEGIN;

DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'mark_page_needs_attention'->'config'
    INTO cfg
    FROM agent_definitions
   WHERE type='content-reviewer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg IS NULL THEN
    RAISE EXCEPTION '370 ROLLBACK: step not found';
  END IF;
  IF cfg ? 'notes_field' OR cfg ? 'validation_issues_field' THEN
    RAISE EXCEPTION '370 ROLLBACK: a key is already present — 370 does not appear applied: %', cfg;
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(
        default_config,
        '{workflow,steps,mark_page_needs_attention,config,notes_field}',
        '"processed_response.rejection_reason"'),
      '{workflow,steps,mark_page_needs_attention,config,validation_issues_field}',
      '"validation_result.issues"'),
    updated_at = now()
WHERE type='content-reviewer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'mark_page_needs_attention'->'config'
    INTO cfg
    FROM agent_definitions
   WHERE type='content-reviewer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF NOT (cfg ? 'notes_field' AND cfg ? 'validation_issues_field') THEN
    RAISE EXCEPTION '370 ROLLBACK VERIFY: restore incomplete: %', cfg;
  END IF;
END $$;

COMMIT;
