-- p4_01b_guide_patents_slot_names.sql — FIX for p4_01: page_components.slot_name was left NULL.
--
-- WHAT WENT WRONG (recorded because it is a cheap mistake to repeat). p4_01 inserted the three
-- sections with component_id set but slot_name NULL. The first section_data_resolved rerender
-- COMPLETED and reported success, but produced nothing:
--   rerender_sections -> {"section_count":3, "rerendered":0, "carried":3}
--   render_page       -> {"skipped":true, "reason":"no components found for page"}
-- i.e. all three sections were CARRIED (stored HTML reused) rather than rendered, and the stored
-- HTML was empty, so assembly found nothing to deploy and the workflow took complete_skipped.
--
-- WHY: rerender_page_sections_action.go:249 looks the component up as `schemas[s.slotName]` —
-- keyed on **slot_name**, NOT component_id (loadStoredSections reads COALESCE(slot_name,'')).
-- A NULL slot_name therefore misses the map and hits the "component not found, carrying stored
-- HTML" branch at :251-257. The lookup key is the component's `function` column, not its `name`:
-- verified against the fleet's existing guide pages —
--   /guides/rng-design/index.html   -> slot_name 'hero', 'generic-text-block'
--   /guides/cma-compliance/index.html -> slot_name 'generic-text-block'
-- and idea.uk's own home page -> 'hero','brief-explanation','tool-list','call-to-action', ...
-- In every case slot_name == content_components.function.
--
-- THE TRAP WORTH NAMING: the orchestration reported COMPLETED. Nothing failed. "Trust the rendered
-- artefact, not the status" (CLAUDE.md / 016b) is exactly this shape — the page did not exist and
-- the job was green.
--
-- FIX: set slot_name from the component's own `function`, so it cannot drift from the lookup key.

\set ON_ERROR_STOP on

BEGIN;

UPDATE page_components pc
SET slot_name = cc.function,
    updated_at = now()
FROM content_components cc, pages p
WHERE cc.id = pc.component_id
  AND p.id = pc.page_id
  AND p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/patents/index.html'
  AND (pc.slot_name IS NULL OR pc.slot_name = '');

DO $guard$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n
  FROM page_components pc JOIN pages p ON p.id = pc.page_id
  WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
    AND p.url = '/guides/patents/index.html'
    AND (pc.slot_name IS NULL OR pc.slot_name = '');
  IF n > 0 THEN
    RAISE EXCEPTION 'ABORT: % section(s) still have no slot_name.', n;
  END IF;
END
$guard$;

COMMIT;

SELECT pc.position, pc.slot_name, cc.name AS component, cc.function
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN content_components cc ON cc.id = pc.component_id
WHERE p.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
  AND p.url = '/guides/patents/index.html'
ORDER BY pc.position;
