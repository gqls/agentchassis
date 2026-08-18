-- 474 — bug 184: turn write-time markdown stripping ON for the two live LLM
-- doors into plain-text content_data fields.
--
--   1. page-content-writer: process_sections_loop.…render_section (action
--      render_component) — strips LLM output at birth, so BOTH surfaces
--      (rendered_html and persisted content_data) are built from clean values.
--   2. section-editor: apply_edit (action apply_section_edit) — strips the
--      merged content map before each branch's render.
--
-- Both flags are read by code shipped with the 184 fix (commit 019fb0616 and
-- the rerender-side follow-up); default OFF in code, so this migration is the
-- entire enablement surface. Step names verified against the LIVE rows
-- 2026-08-18 (seed-vs-live drift is a known trap): page-content-writer's step
-- is render_section inside the sub_workflow; section-editor's step is
-- apply_edit.
--
-- NOT touched, deliberately (migration 304's own measurement, re-stated):
-- content-writer and simple-content-writer-with-approval carry neither the
-- STRICT RULES prompt nor the save_page_sections path; tool writers
-- legitimately ask for markdown-fenced JSON (227) and do not write plain-text
-- content_data through these seams.
--
-- ORDERING: safe before the image (keys unread by the old binary); intended
-- to be applied together with 473 after the image is live.

BEGIN;

CREATE TABLE IF NOT EXISTS _backup_474_writer_strip AS
  SELECT id, type, default_config, now() AS backed_up_at
    FROM agent_definitions
   WHERE type IN ('page-content-writer', 'section-editor') AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,strip_literal_markdown}',
         'true'::jsonb)
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   -- anchor: the step must exist with its expected action, or refuse
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,action}' = 'render_component';

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,apply_edit,config,strip_literal_markdown}',
         'true'::jsonb)
 WHERE type = 'section-editor' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,apply_edit,action}' = 'apply_section_edit';

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND (
       (type = 'page-content-writer' AND
        (default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,strip_literal_markdown}')::text = 'true')
       OR
       (type = 'section-editor' AND
        (default_config #> '{workflow,steps,apply_edit,config,strip_literal_markdown}')::text = 'true')
     );
  IF n <> 2 THEN
    RAISE EXCEPTION '474 FAILED: expected both agents flagged, got % — a step path anchor has moved; read the live rows and re-anchor', n;
  END IF;
  RAISE NOTICE '474 OK: both agents flagged';
END $$;

COMMIT;
