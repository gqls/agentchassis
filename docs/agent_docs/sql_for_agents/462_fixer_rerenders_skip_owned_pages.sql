-- 462 — component-template-fixer's create_rerender skips rebuild_policy='owned'
--       pages (bugs_open/283 §13.6; found by the batch, 2026-08-18)
--
-- 460 made the fixer file page-scoped template_changed rerenders. The batch
-- then filed 17 of them against tool-OWNED pages, where save_page_sections
-- correctly refuses a generic section save ("a generic section save would
-- clobber it") — a deterministic refusal: 2 failed before the pattern was
-- diagnosed, 15 were cancelled post-diagnosis. An owned page's conversion
-- reaches it when its OWNING pipeline next renders, or via a targeted
-- apply_section_edit (which binds InstanceID); a generic rerender never will.
-- So the fixer stops filing items that exist only to be refused.
--
-- Guarded on 461's post-image; verify PREPARE-compiles the embedded query
-- (the check class 460 lacked — LANDMINES 2026-08-18).
BEGIN;

SELECT snapshot_agent('component-template-fixer', '462 pre-image: skip owned pages in create_rerender');

UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,create_rerender,config,query}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,create_rerender,config,query}',
        'WHERE pc.component_id = $1::uuid AND NOT EXISTS',
        'WHERE pc.component_id = $1::uuid AND p.rebuild_policy IS DISTINCT FROM ''owned'' AND NOT EXISTS'
      ))
    ),
    updated_at = now()
WHERE type = 'component-template-fixer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,create_rerender,config,query}'
      LIKE '%WHERE pc.component_id = $1::uuid AND NOT EXISTS%'
  AND default_config #>> '{workflow,steps,create_rerender,config,query}'
      NOT LIKE '%rebuild_policy%';

DO $$
DECLARE q text;
BEGIN
  SELECT default_config #>> '{workflow,steps,create_rerender,config,query}'
    INTO q FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF q NOT LIKE '%rebuild_policy IS DISTINCT FROM ''owned''%' THEN
    RAISE EXCEPTION '462: owned-page exclusion not present: %', left(q,200);
  END IF;
  IF q NOT LIKE '%template_changed%' THEN
    RAISE EXCEPTION '462: 460/461 intent lost: %', left(q,160);
  END IF;
  EXECUTE 'PREPARE chk462 AS ' || q;
  DEALLOCATE chk462;
END $$;

COMMIT;
