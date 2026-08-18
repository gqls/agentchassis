-- 474 — bug 184: turn write-time markdown stripping ON for the two live LLM
-- doors into plain-text content_data fields.
--
--   1. page-content-writer: process_sections_loop.…render_section (action
--      render_component) — strips LLM output at birth, so BOTH surfaces
--      (rendered_html and persisted content_data) are built from clean values.
--   2. section-editor: apply_edit (action apply_section_edit) — strips the
--      merged content map before each branch's render.
--
-- Both flags are read from params.StepConfig.Config — RenderComponentAction's
-- `config := params.StepConfig.Config` (v3_site_actions.go:1836) and
-- ApplySectionEditAction's switch-site read — which is where the engine
-- delivers these jsonb paths (coordinator.go:1696). Default OFF in code, so
-- this migration is the entire enablement surface. Step paths verified against
-- the LIVE rows 2026-08-18, and re-verified by council round 1's own read-only
-- checks (corr 060bcc0a): page-content-writer's sub_workflow uses key 'steps'
-- with step 'render_section'; section-editor's step is 'apply_edit'. The
-- UPDATEs are additionally anchored on each step's action value, so a moved
-- path means 0 rows and a loud RAISE — jsonb_set can never mint an orphan key
-- on a row the anchor rejected.
--
-- NOT touched, deliberately (migration 304's own measurement, re-stated):
-- content-writer and simple-content-writer-with-approval carry neither the
-- STRICT RULES prompt nor the save_page_sections path; tool writers
-- legitimately ask for markdown-fenced JSON (227) and do not write plain-text
-- content_data through these seams.
--
-- ORDERING: safe before the image (keys unread by the old binary); intended
-- to be applied together with 473 after the image is live.
--
-- Backup: snapshot_agent() per row (standard idiom, per council round 1 reuse
-- objection). Needle-gated: a re-run where the flag is already true updates 0
-- rows; the verify checks final state, so re-runs pass without lying.
--
-- NULL-DIRECTION ANALYSIS of the verify (council r3 — the jsonb <>-vs-NULL
-- landmine): the DO block counts POSITIVE presence (`(#> path)::text = 'true'`);
-- an absent path yields NULL, the row is not counted, n <> 2, RAISE. No
-- negative-form (`<>`) comparison exists here, so no NULL can read as green.
-- A jsonb_set path typo cannot no-op silently either: jsonb_set with a missing
-- parent returns its input unchanged, the flag then does not exist at the
-- checked path, and the same count check RAISES. (Per-UPDATE row-count
-- assertions are deliberately absent: the needle gates make 0-row re-runs
-- legitimate, and the final-state check is the idempotency-correct verify.)

BEGIN;

SELECT snapshot_agent('page-content-writer',
                      '474_writer_render_strip_optin.sql: pre-update');
SELECT snapshot_agent('section-editor',
                      '474_writer_render_strip_optin.sql: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,strip_literal_markdown}',
         'true'::jsonb),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   -- anchor: the step must exist with its expected action, or refuse
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,action}' = 'render_component'
   -- needle gate: a re-run is a 0-row no-op
   AND (default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,render_section,config,strip_literal_markdown}')::text
       IS DISTINCT FROM 'true';

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,apply_edit,config,strip_literal_markdown}',
         'true'::jsonb),
       updated_at = now()
 WHERE type = 'section-editor' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,apply_edit,action}' = 'apply_section_edit'
   AND (default_config #> '{workflow,steps,apply_edit,config,strip_literal_markdown}')::text
       IS DISTINCT FROM 'true';

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
