-- 162_toolgen_enqueue_rerender_tail.sql — close the tool-birth deploy gap.
--
-- tool-generator built a component + page + nav + PLAN but NEVER enqueued a
-- page_rerender, so every newly born tool page sat build_status='planned'
-- until an unrelated sweep happened to pick it up. All three births to date
-- (xp-curve, drop-rate, loot-table) needed a hand-inserted work item. This
-- adds a tail step that enqueues the rerender itself, using create_rerender_
-- items' new SINGLE-PAGE mode (page_id/name/filename scalar config) fed from
-- save_tool's create_result.
--
-- Chain change: index_plan.next_step  complete -> enqueue_rerender -> complete.
--
-- SAFE ON THE CURRENT IMAGE, ACTIVATES ON THE NEXT (the 147 pattern): the
-- deployed create_rerender_items has no single-page mode, so with no
-- rerender_pages.pages array it returns "no pages found" (items_created 0) —
-- a harmless no-op, exactly today's behaviour. Once the image carrying
-- singlePageFromScalars ships, the same step enqueues the rerender for real.
-- error_step=complete means a doc/deploy tail can never fail the tool build
-- (the standing "docs never fail the work" rule).
--
-- Applied out of band (psql -f + ledger row same sitting per aaa_fails_to_mend/007).

BEGIN;

SELECT snapshot_agent('tool-generator', '162_toolgen_enqueue_rerender_tail: pre-update');

DO $$
DECLARE
  cfg jsonb;
  new_step jsonb := $step${
    "action": "create_rerender_items",
    "description": "Enqueue the page_rerender for the just-created tool page so it deploys without waiting for a sweep (single-page mode; the tool-birth deploy gap). No-op on images without single-page create_rerender_items.",
    "config": {
      "site_id": "site_record.site_id",
      "domain": "input_data.domain",
      "page_id_field": "create_result.page_id",
      "page_name_field": "create_result.function",
      "filename_field": "create_result.page_url",
      "error_step": "complete"
    },
    "next_step": "complete",
    "output_field": "rerender_enqueue"
  }$step$::jsonb;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
  WHERE type='tool-generator' AND is_active;
  IF cfg IS NULL THEN
    RAISE EXCEPTION '162: no active tool-generator';
  END IF;

  -- Guards: we are rerouting the version we inspected.
  IF cfg #>> '{workflow,steps,index_plan,next_step}' <> 'complete' THEN
    RAISE EXCEPTION '162: index_plan.next_step is not "complete" (got %) — re-inspect before rerouting',
      cfg #>> '{workflow,steps,index_plan,next_step}';
  END IF;
  IF cfg #> '{workflow,steps,enqueue_rerender}' IS NOT NULL THEN
    RAISE EXCEPTION '162: enqueue_rerender step already exists';
  END IF;

  cfg := jsonb_set(cfg, '{workflow,steps,enqueue_rerender}', new_step);
  cfg := jsonb_set(cfg, '{workflow,steps,index_plan,next_step}', '"enqueue_rerender"'::jsonb);

  UPDATE agent_definitions SET default_config = cfg
  WHERE type='tool-generator' AND is_active;
END $$;

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'build',
  '## tool-generator now enqueues its own page_rerender
Observed: newly created tool pages stayed build_status=planned until an unrelated sweep rendered them — tool-generator never enqueued a page_rerender (all 3 births to date needed a hand-inserted item; create_tool_component even returned an unused needs_rerender:true).
Root cause: no tail step; the enqueue was never wired.
Fix: index_plan -> enqueue_rerender -> complete, using create_rerender_items single-page mode fed from create_result (migration 162). Snapshot taken. Safe/no-op on the current image; activates on the image carrying single-page create_rerender_items; error_step=complete so a deploy tail can never fail the build.
Verified: guarded reroute applied; live activation gated on the next chassis image.
Categories: fix',
  '["fix"]'::jsonb,
  'migration', '162_toolgen_enqueue_rerender_tail'
);

COMMIT;
