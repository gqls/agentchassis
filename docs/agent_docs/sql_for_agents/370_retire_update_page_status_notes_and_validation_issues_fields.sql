-- ============================================================================
-- 370_retire_update_page_status_notes_and_validation_issues_fields.sql
--
-- Increment #5 of the RemovedConfigKeys adoptions (RFC_021 Q3, owner ruling
-- 2026-08-10: "proceed as proposed"), executed under the Q1 protocol
-- (census-at-adoption + the removed-config-keys-check CronJob).
--
-- The two keys were adjudicated DEAD by migration 356's enumeration:
-- UpdatePageStatusAction (v3_site_actions.go) reads exactly five config keys —
-- status, page_id_field, site_id_field, page_name_field,
-- page_component_id_field — with no extractor and no registered spec, so
-- `notes_field` and `validation_issues_field` are read by nothing. 356 left
-- them standing because, unlike commit_from, they encode an author's INTENT the
-- action has never had: recording WHY a page was flagged
-- (notes_field = processed_response.rejection_reason,
-- validation_issues_field = validation_result.issues) — and pages has no
-- column for it. The council's bug_historian seat then objected (corr
-- 3eb0d1f1, round 1) that "left standing, noted in prose" is untracked
-- deferral; the owner's Q3 ruling resolves it: THE INTENT IS RECORDED, THE
-- KEYS ARE RETIRED.
--
-- Where the intent now lives (so deleting the keys erases nothing):
--   * this header;
--   * migration 356's header (the original adjudication);
--   * the RemovedConfigKeys message on update_page_status (shipped in the same
--     commit), which every future author who writes the key again will be
--     shown verbatim by the validator.
-- If the platform ever wants the behaviour, that is a feature: a place for
-- review notes to live (a pages column or a doc_notes convention), built on
-- its own merits — not a config key silently resolving to nowhere.
--
-- Census at adoption (Q1 protocol), 2026-08-10: exactly ONE live step carries
-- either key on this action — content-reviewer.mark_page_needs_attention, at
-- the top level. (`validation_issues_field` also appears on
-- content-reviewer.escalate_to_human — action `request_human_input`, a
-- DIFFERENT action, out of this adjudication's scope and untouched here.)
--
-- ORDERING: this migration MUST be applied before the Go commit declaring the
-- keys removed — the declaring binary rejects any carrier on every message.
-- Same rule as 364; on this tree committing is shipping.
--
-- Seed corrected in the same commit: 025_content_reviewer_agent.sql (§9).
-- Live immediately (DB config; no image roll involved).
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
    RAISE EXCEPTION '370: content-reviewer.mark_page_needs_attention not found';
  END IF;
  IF NOT (cfg ? 'notes_field') AND NOT (cfg ? 'validation_issues_field') THEN
    RAISE EXCEPTION '370: already applied (neither key present)';
  END IF;
  IF cfg->>'notes_field' IS DISTINCT FROM 'processed_response.rejection_reason'
     OR cfg->>'validation_issues_field' IS DISTINCT FROM 'validation_result.issues'
     OR cfg->>'status' IS DISTINCT FROM 'needs_attention'
     OR cfg->>'page_id_field' IS DISTINCT FROM 'input_data.current_page.id' THEN
    RAISE EXCEPTION '370: DRIFT — config is not the expected shape: %. Re-measure.', cfg;
  END IF;
END $$;

UPDATE agent_definitions
SET default_config = default_config
      #- '{workflow,steps,mark_page_needs_attention,config,notes_field}'
      #- '{workflow,steps,mark_page_needs_attention,config,validation_issues_field}',
    updated_at = now()
WHERE type='content-reviewer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Verify: removal AND sibling survival (whole-config-clobber check), DO/RAISE.
DO $$
DECLARE cfg jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'mark_page_needs_attention'->'config'
    INTO cfg
    FROM agent_definitions
   WHERE type='content-reviewer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF cfg ? 'notes_field' OR cfg ? 'validation_issues_field' THEN
    RAISE EXCEPTION '370 VERIFY: a retired key survived: %', cfg;
  END IF;
  IF cfg->>'status' IS DISTINCT FROM 'needs_attention'
     OR cfg->>'page_id_field' IS DISTINCT FROM 'input_data.current_page.id' THEN
    RAISE EXCEPTION '370 VERIFY: a sibling key was lost: %', cfg;
  END IF;
  -- escalate_to_human's validation_issues_field (a DIFFERENT action) must be
  -- untouched — this migration's scope is one step.
  IF NOT (
    SELECT default_config->'workflow'->'steps'->'escalate_to_human'->'config' ? 'validation_issues_field'
      FROM agent_definitions
     WHERE type='content-reviewer' AND is_active
       AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL) THEN
    RAISE EXCEPTION '370 VERIFY: escalate_to_human lost its validation_issues_field — out-of-scope edit';
  END IF;
END $$;

COMMIT;
