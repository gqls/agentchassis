-- 750: copy-editor's approval fans out — one section_edit per proposed edit.
--
-- bugs_open/466. The config half of the fix. The Go half is in
-- internal/core-manager/admin/site_admin_handlers.go (HandleApproveWorkItem).
--
-- ⚠ ORDERING: this file is INERT until a binary carrying the fan-out rolls, and
-- HARMLESS before it — an older admin binary simply does not read `fan_out_from`
-- or `defaults`, and files the same single (broken) follow-on it files today. So
-- there is no bad window and this is NOT a _HOLD file. It names no action, so
-- the "a seed naming an unregistered action fails at runtime" rule does not
-- bite either. It just does nothing until the roll.
--
-- WHAT WAS WRONG, measured rather than argued:
--
--   * `include_fields: ["copy_edit","page_target"]` could never resolve. The
--     approve handler looked those names up in the REVIEW ITEM'S SPEC, and
--     checkpoint_for_review — the only producer of these items — writes a fixed
--     key set (review_data, checkpoint, source_agent, correlation_id, domain?,
--     spec_aspect?, on_approve). [MEASURED 2026-09-03] 42 field mentions across
--     21 items, all history since 2026-08-24: ZERO present at spec top level.
--     `domain` replaces them because `domain` is a key the spec genuinely holds,
--     and it is what the two proven hand-filed section_edit items carried.
--
--   * The shapes did not match even once the names were plumbed: copy-editor
--     proposes N edits, section-editor applies ONE. `fan_out_from: "edits"`
--     files one child per element, merging that element's own fields at the top
--     of the child spec — which is where load_edit_context and apply_section_edit
--     read them.
--
--   * `defaults` supplies what the proposal does not carry. [MEASURED
--     2026-09-03] of the 41 proposed edits in those 21 items, 41 carry
--     `page_component_id`, ZERO carry `edit_type` and ZERO carry `page_name`.
--     So `edit_type` must be defaulted, and `page_name` must NOT be — every edit
--     is addressed by page_component_id, which alone satisfies load_edit_context,
--     and a defaulted page_name would be a guess applied to every page.
--
-- ROLLBACK: 750_..._ROLLBACK.sql restores the previous on_approve verbatim.

BEGIN;

-- ── snapshot ────────────────────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS bak_750_copy_editor_20260903 AS
SELECT * FROM agent_definitions
WHERE type = 'copy-editor'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

-- ── anchor guard: refuse if the step is not the shape this file was written
--    against. Another session may have changed it since. ─────────────────────
DO $$
DECLARE
  n int;
  cur jsonb;
BEGIN
  SELECT count(*) INTO n
  FROM agent_definitions
  WHERE type = 'copy-editor' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '750: expected exactly 1 live copy-editor row, found %', n;
  END IF;

  SELECT default_config->'workflow'->'steps'->'request_review'->'config'->'on_approve'
    INTO cur
  FROM agent_definitions
  WHERE type = 'copy-editor' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cur IS NULL THEN
    RAISE EXCEPTION '750: request_review.config.on_approve is absent — the step moved';
  END IF;
  IF cur->>'item_type' <> 'section_edit' OR cur->>'handler_agent' <> 'section-editor' THEN
    RAISE EXCEPTION '750: on_approve targets %/% , expected section_edit/section-editor',
      cur->>'item_type', cur->>'handler_agent';
  END IF;
  IF cur ? 'fan_out_from' THEN
    RAISE EXCEPTION '750: fan_out_from already present — already applied, or another session got here first';
  END IF;
END $$;

-- ── the change ──────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,request_review,config,on_approve}',
      jsonb_build_object(
        'item_type',      'section_edit',
        'handler_agent',  'section-editor',
        'fan_out_from',   'edits',
        'defaults',       jsonb_build_object('edit_type', 'content_edit'),
        'include_fields', jsonb_build_array('domain')
      ),
      false
    ),
    updated_at = NOW()
WHERE type = 'copy-editor'
  AND is_active
  AND COALESCE(is_snapshot, false) = false
  AND deleted_at IS NULL;

-- ── verify: a DO/RAISE block, because ON_ERROR_STOP does NOT abort a COMMIT on
--    a SELECT that merely returns the wrong rows (RFC_006's landmine). ───────
DO $$
DECLARE
  cur jsonb;
BEGIN
  SELECT default_config->'workflow'->'steps'->'request_review'->'config'->'on_approve'
    INTO cur
  FROM agent_definitions
  WHERE type = 'copy-editor' AND is_active
    AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF cur->>'fan_out_from' <> 'edits' THEN
    RAISE EXCEPTION '750 VERIFY: fan_out_from = %, expected edits', cur->>'fan_out_from';
  END IF;
  IF cur->'defaults'->>'edit_type' <> 'content_edit' THEN
    RAISE EXCEPTION '750 VERIFY: defaults.edit_type = %, expected content_edit', cur->'defaults'->>'edit_type';
  END IF;
  IF cur->'defaults' ? 'page_name' THEN
    RAISE EXCEPTION '750 VERIFY: defaults must NOT carry page_name — every edit is addressed by page_component_id';
  END IF;
  IF cur->'include_fields' <> jsonb_build_array('domain') THEN
    RAISE EXCEPTION '750 VERIFY: include_fields = %, expected ["domain"]', cur->'include_fields';
  END IF;
  -- the parts that must be UNDISTURBED
  IF cur->>'item_type' <> 'section_edit' OR cur->>'handler_agent' <> 'section-editor' THEN
    RAISE EXCEPTION '750 VERIFY: the follow-on target was disturbed: %', cur;
  END IF;
  RAISE NOTICE '750: copy-editor approval fans out — one section_edit per edit, edit_type defaulted, domain carried';
END $$;

COMMIT;
