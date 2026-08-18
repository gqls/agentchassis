-- 460 — 'template_changed' joins the page-rerender reason vocabulary, and
--       component-template-fixer files PAGE-SCOPED rerenders that carry it
--       (bugs_open/283 §13; found by the canary, 2026-08-18)
--
-- THE GAP, measured end-to-end on the first live conversion. The 283 canary
-- converted tool-css-unit-converter's template (fixed:true, every count as
-- predicted), the fixer's create_rerender raised its needs_rerender, the
-- rerender-pages fan-out created 111 page_rerender items, all 111 completed,
-- the page DEPLOYED — and the served page still carries the OLD ids, because
-- page-rerender's check_rerender_mode routes to the sections re-render ONLY for
--   image_landed | section_data_resolved | cta_links_stale
-- and everything else ASSEMBLES STORED HTML. "The component's template changed"
-- has no reason in the vocabulary at all, so every template fix ships stale
-- bytes with a green status. This is CLC-002's own recorded history repeating:
-- "pre-fix, that item carried no reason, making the triggered re-render
-- assemble-only and unable to repair anything."
--
-- TWO CHANGES, one per agent row, each guarded on a VERBATIM pre-image so a
-- concurrent edit by another session ABORTS this file rather than being
-- silently reverted:
--
--  1. page-rerender.check_rerender_mode gains
--       OR input_data.spec.reason == 'template_changed'
--     Additive: the three existing reasons and the else-branch are unchanged,
--     so no currently-flowing item changes behaviour. Per the 2026-07-29 owner
--     ruling this is a vocabulary addition that changes no guarantee (the
--     sections path already exists for three reasons; this names a fourth
--     cause) — normal council gate, not an RFC.
--
--  2. component-template-fixer.create_rerender is REPLACED: instead of one
--     site-wide, reason-less needs_rerender (which fans out to EVERY page and
--     re-renders NONE of them from templates), it files one page_rerender item
--     PER PAGE CARRYING THE FIXED COMPONENT, with spec.reason='template_changed',
--     across every site the component is placed on (the cross-domain row of
--     RFC_034 §1 is covered by taking site_id from the page, not the item).
--     Spec shape copied from the live rerender-pages fan-out items
--     (page_id/page_name/domain/filename) so page-rerender's input mapping
--     needs nothing new. Dedup: NOT EXISTS on an open template_changed
--     page_rerender for the same page.
--
-- Apply-time verify is DO/RAISE (a SELECT-only verify cannot stop the COMMIT).
BEGIN;

-- Snapshots first (agent_definitions_backup, the estate's own function).
SELECT snapshot_agent('page-rerender', '460 pre-image: template_changed reason');
SELECT snapshot_agent('component-template-fixer', '460 pre-image: page-scoped template_changed rerenders');

-- ── Change 1: the vocabulary ────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,check_rerender_mode,config,condition}',
      to_jsonb(
        (default_config #>> '{workflow,steps,check_rerender_mode,config,condition}')
        || ' OR input_data.spec.reason == ''template_changed'''
      )
    ),
    updated_at = now()
WHERE type = 'page-rerender' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
      = 'input_data.spec.reason == ''image_landed'' OR input_data.spec.reason == ''section_data_resolved'' OR input_data.spec.reason == ''cta_links_stale''';

-- ── Change 2: the fixer files page-scoped rerenders that carry the reason ───
UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(
        default_config,
        '{workflow,steps,create_rerender,config,query}',
        to_jsonb(
          'INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, priority, handler_agent, status, created_by, spec) '
          || 'SELECT p.site_id, ''side_effect'', ''build'', ''page_rerender'', ''low'', '
          || '''Rerender page after template fix: '' || p.name, 80, ''page-rerender'', ''triaged'', ''component-template-fixer'', '
          || 'jsonb_build_object(''reason'',''template_changed'',''page_id'',p.id::text,''page_name'',p.name,''domain'',s.domain,''filename'',p.filename) '
          || 'FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id '
          || 'WHERE pc.component_id = $1::uuid '
          || 'AND NOT EXISTS (SELECT 1 FROM site_work_items w WHERE w.site_id = p.site_id '
          || 'AND w.item_type = ''page_rerender'' AND w.spec->>''page_id'' = p.id::text '
          || 'AND w.spec->>''reason'' = ''template_changed'' '
          || 'AND w.status IN (''detected'',''triaged'',''claimed''))'
        )
      ),
      '{workflow,steps,create_rerender,config,params}',
      '["fix_result.component_id"]'::jsonb
    ),
    updated_at = now()
WHERE type = 'component-template-fixer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,create_rerender,config,query}'
      LIKE 'INSERT INTO site_work_items%needs_rerender%rerender-pages%';

-- ── Verify, or refuse the COMMIT ────────────────────────────────────────────
DO $$
DECLARE cond text; q text; prm text;
BEGIN
  SELECT default_config #>> '{workflow,steps,check_rerender_mode,config,condition}'
    INTO cond FROM agent_definitions
   WHERE type='page-rerender' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cond NOT LIKE '%template_changed%' THEN
    RAISE EXCEPTION '460: page-rerender condition not updated (pre-image guard missed — concurrent edit?): %', cond;
  END IF;
  IF cond NOT LIKE '%image_landed%' OR cond NOT LIKE '%section_data_resolved%' OR cond NOT LIKE '%cta_links_stale%' THEN
    RAISE EXCEPTION '460: an existing reason vanished from the condition: %', cond;
  END IF;

  SELECT default_config #>> '{workflow,steps,create_rerender,config,query}',
         default_config #>> '{workflow,steps,create_rerender,config,params}'
    INTO q, prm FROM agent_definitions
   WHERE type='component-template-fixer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF q NOT LIKE '%template_changed%' OR q NOT LIKE '%page_rerender%' THEN
    RAISE EXCEPTION '460: fixer create_rerender not updated: %', left(q,120);
  END IF;
  IF prm NOT LIKE '%fix_result.component_id%' THEN
    RAISE EXCEPTION '460: fixer create_rerender params not updated: %', prm;
  END IF;
END $$;

COMMIT;
