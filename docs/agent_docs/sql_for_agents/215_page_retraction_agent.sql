-- 215_page_retraction_agent.sql — seat `page-retraction`, the caller for the
-- UNPUBLISH primitive (bugs_open/098, concept register DGH-006).
--
-- WHY THIS EXISTS AT ALL, rather than leaving the action unreachable:
-- the council's round-1 review (correlation 4a7f0877) objected, correctly, that
-- the plan "builds the delete_file/retract capability but includes no edit that
-- actually invokes it against the cited live cases". This platform has been
-- bitten before by a mechanism that ships and then rots unexercised, and the
-- 2026-07-29 owner ruling declined to require default-OFF switches for exactly
-- that reason. So the capability gets a caller in the same push.
--
-- IT IS DELIBERATELY NOT SCHEDULED AND NOT WIRED INTO THE ARCHIVE PATH.
-- Nothing in the codebase archives a page — there is no writer of
-- pages.status='archived' in Go or in any frontend; archiving is a hand-run SQL
-- operation. So retraction is operator-initiated by construction, and this row
-- is what an operator (or a future archive handler, or a work-item handler)
-- dispatches. Making it a cron would mean the platform started deleting live
-- files on a timer off the back of a hand-edited column, which is not a thing
-- anyone asked for.
--
-- SAFETY IS IN THE ACTION, NOT IN THIS CONFIG, and that is on purpose: a guard
-- that lives in workflow JSON can be forgotten by the next author of the next
-- workflow. retract_page_deployment refuses a page whose url names no file of
-- its own, refuses a path an ACTIVE page also derives, refuses a page still
-- linked from body copy or site chrome, retires nav rows pointing at it, and
-- reports every refusal rather than swallowing it. This row cannot switch any
-- of that off — there is no config key for it.
--
-- DISPATCH (site_id is required; page_ids optional — omit it to consider every
-- non-active page on the site that still carries a deploy stamp):
--   input_data: {"site_id": "<uuid>", "page_ids": ["<uuid>", ...]}
-- Set "dry_run": true in the step config to audit without deleting.

INSERT INTO agent_definitions (type, name, description, default_config, is_active)
VALUES (
  'page-retraction',
  'Page Retraction',
  'UNPUBLISH: removes the deployed artefacts of pages the platform no longer wants served, retires their nav rows, and reports anything stranded. Refuses a page still linked from live content.',
  jsonb_build_object(
    'workflow', jsonb_build_object(
      'start_step', 'retract',
      'steps', jsonb_build_object(
        'retract', jsonb_build_object(
          'action', 'retract_page_deployment',
          'description', 'Audit the retraction graph, retire nav rows, dispatch delete_file',
          'config', jsonb_build_object(
            'site_id_field',  'input_data.site_id',
            'page_ids_field', 'input_data.page_ids'
          ),
          'output_field', 'retraction',
          'next_step', 'complete'
        ),
        'complete', jsonb_build_object(
          'action', 'complete_workflow',
          'description', 'Retraction complete',
          'config', jsonb_build_object('output_fields', jsonb_build_array('retraction'))
        )
      )
    )
  ),
  true
)
ON CONFLICT DO NOTHING;

-- Verification (run after applying):
--   SELECT type, is_active,
--          default_config #>> '{workflow,steps,retract,action}' AS action
--     FROM agent_definitions
--    WHERE type='page-retraction' AND is_active
--      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- expect exactly one row, action = retract_page_deployment
