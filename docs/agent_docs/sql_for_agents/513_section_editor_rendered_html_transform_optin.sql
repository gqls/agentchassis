-- 513 — bugs_open/277 §5 (owner ruling 2026-08-20: the seven rendered_html
-- code_span findings DO get a repair route): enable the rendered_html_transform
-- edit type on section-editor's apply_edit step, and let the step extract the
-- transform_name input.
--
-- WHAT THE FLAG GATES. apply_section_edit (section_editor_actions.go) gained a
-- third edit_type, rendered_html_transform: apply a NAMED deterministic
-- transform (code_span_to_code_tag — datahelpers.ConvertLiteralCodeSpansInHTML,
-- `x` → <code>x</code> in assertion text, byte-spliced, detector's skip set) to
-- the component's EXISTING rendered_html, content_data untouched. It exists for
-- components whose content_data cannot reproduce their rendered_html (Ported
-- Page: 100 of 115 instances hold NONE of their template's fields), which are
-- unreachable by every regenerate-from-source route BY CONSTRUCTION — the
-- LANDMINES entry "A component whose content_data CANNOT REPRODUCE…" holds the
-- evidence. Default OFF in code per the 2026-08-02 §2 ruling (new authority on
-- a shared seam ships as an opt-in whose unsafe default is OFF); this migration
-- is the entire enablement surface, and section-editor's apply_edit is the only
-- step it enables.
--
-- TWO EDITS, ONE STEP:
--   1. config.allow_rendered_html_transform = true   (the authority)
--   2. config.input_fields += "transform_name"        (the plumbing: input_fields
--      is ExtractActionInputs' Strategy-1 WHITELIST — action_inputs.go:831 —
--      so an input absent from it is not extracted at all)
--
-- ORDERING against the chassis roll carrying the code (both directions safe,
-- stated because a session WILL ask):
--   - applied BEFORE the roll: the old binary reads neither key — inert. No
--     item carries the new spec shape until the new DETECTOR
--     (check_literal_markdown's transformRouteSlot) rolls in the same image
--     as the action branch, so nothing can dispatch against half a deploy.
--   - applied AFTER the roll: a new-shape item dispatched in the window is
--     refused loudly by the action's config gate, fails into the attempt
--     ladder, and succeeds on a retry once this is applied. Self-healing.
--
-- THE PAIR STAYS HELD EITHER WAY: literal_markdown → section-editor has zero
-- lifetime completes, so the 444 promoter's ≥1-complete door holds it until a
-- canary run completes (the deliberate bootstrap — bugs_open/277 records it).
-- This migration changes WHICH work is offered to the promoter, not whether it
-- promotes.
--
-- Backup: snapshot_agent() per row (474's idiom). Needle-gated: a re-run where
-- the flag is already true / the field already listed updates 0 rows; the
-- verify checks final state, so re-runs pass without lying.
--
-- NULL-DIRECTION ANALYSIS of the verify (the jsonb <>-vs-NULL landmine): both
-- DO-block checks count POSITIVE presence; an absent path yields NULL, the row
-- is not counted, n <> 1, RAISE. No negative-form comparison exists here.

BEGIN;

SELECT snapshot_agent('section-editor',
                      '513_section_editor_rendered_html_transform_optin.sql: pre-update');

-- 1. The authority: allow_rendered_html_transform = true, anchored on the
--    step's action value (a moved step means 0 rows and the verify RAISEs).
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,apply_edit,config,allow_rendered_html_transform}',
         'true'::jsonb),
       updated_at = now()
 WHERE type = 'section-editor' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,apply_edit,action}' = 'apply_section_edit'
   AND (default_config #> '{workflow,steps,apply_edit,config,allow_rendered_html_transform}')::text
       IS DISTINCT FROM 'true';

-- 2. The plumbing: append "transform_name" to the step's input_fields
--    whitelist. Anchored the same way; needle-gated on the value not already
--    being present (jsonb @> on the array).
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,apply_edit,config,input_fields}',
         (default_config #> '{workflow,steps,apply_edit,config,input_fields}') || '["transform_name"]'::jsonb),
       updated_at = now()
 WHERE type = 'section-editor' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,apply_edit,action}' = 'apply_section_edit'
   AND jsonb_typeof(default_config #> '{workflow,steps,apply_edit,config,input_fields}') = 'array'
   AND NOT (default_config #> '{workflow,steps,apply_edit,config,input_fields}') @> '["transform_name"]'::jsonb;

-- Verify. RAISE, not SELECT — a plain SELECT cannot stop the COMMIT.
DO $$
DECLARE n int;
BEGIN
  -- Positive presence of the flag on exactly the one live row.
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'section-editor' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND (default_config #> '{workflow,steps,apply_edit,config,allow_rendered_html_transform}')::text = 'true';
  IF n <> 1 THEN
    RAISE EXCEPTION '513: allow_rendered_html_transform is true on % live section-editor rows, expected exactly 1', n;
  END IF;

  -- Positive presence of the input_fields entry, exactly once (a double
  -- append would mean the needle gate failed — count elements, not rows).
  SELECT count(*) INTO n
    FROM agent_definitions ad,
         jsonb_array_elements_text(ad.default_config #> '{workflow,steps,apply_edit,config,input_fields}') f
   WHERE ad.type = 'section-editor' AND ad.is_active
     AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
     AND f = 'transform_name';
  IF n <> 1 THEN
    RAISE EXCEPTION '513: input_fields carries % ''transform_name'' entries, expected exactly 1', n;
  END IF;

  RAISE NOTICE '513 OK: rendered_html_transform enabled on section-editor.apply_edit (flag true, transform_name whitelisted once).';
END $$;

COMMIT;
