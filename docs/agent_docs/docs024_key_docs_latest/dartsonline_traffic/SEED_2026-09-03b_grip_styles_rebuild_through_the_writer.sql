-- SEED_2026-09-03b_grip_styles_rebuild_through_the_writer.sql
--
-- STAGE 3 of the grip-styles recompose. Stage 1 wrote the 11-section plan, the subjects and the
-- five section-bound figures (`SEED_2026-09-03`); stage 2 generated the five illustrations. This
-- files the rebuild that makes the page actually BE its new composition.
--
-- WHY A `needs_content_page` AND NOT A RE-RENDER. The page's prose currently lives in ONE
-- `article-body` component whose field is documented "write the full article body as HTML". The
-- new composition has nine prose sections, each with its own subject. **Only the writer can
-- redistribute that**; a re-render assembles stored `content_data` and would redeploy the same
-- three sections. This is also IMG-075's own grading rule: a re-render shows that nothing broke,
-- **the save path is the decisive test**, because surviving a save is the property the whole
-- per-section mechanism exists for.
-- Shape copied from a worked example applied today on another lane (finetuning.uk `/index.html`,
-- `check='gap_plan_new_page'`, `sections` + `suggestion`), not invented here.
--
-- ⚠ WHY THIS CANNOT CLOBBER STAGE 1, CHECKED BEFORE FILING RATHER THAN HOPED. The only action
-- that writes `site_plan_sections` on a build path is `ensure_page_section_layout_action.go`
-- (calling `insertSitePlanSectionRows`, whose SQL lives in `apply_gap_plan_action.go`). It is
-- **fill-only and refuses twice**: once if `pages.sections` is non-empty — *"this action never
-- overwrites an existing layout"* — and again if `count(*) FROM site_plan_sections` for the page
-- is non-zero: *"current plan already carries N section row(s) for this page — refusing"*.
-- grip-styles now has 11. So the subjects and the ordinals survive the rebuild, which matters
-- because `insertSitePlanSectionRows` does not carry a `subject` column at all and would have
-- silently dropped every one of them.
--
-- ⚠ `pages.sections` IS STILL THE OLD THREE, AND THAT IS EXPECTED. It is the materialised cache
-- (tier 3); `load_page_sections_from_spec_action` resolves the plan tables FIRST (tier 1) and
-- syncs down to the cache once it resolves. The `sections` array below is therefore stated to
-- match the PLAN, not the cache — if the two disagree at bind time, `sectionOrderAgrees` stands
-- the whole per-section binding down by design, and the page would build with page-wide imagery
-- and look entirely correct. **That is the failure to watch for at verification, and its tell is
-- five sections sharing one image rather than any error.**
--
-- RESTORE POINTS: `bak_20260903_gripstyles_components` (3 rows, article-body content_data 8,401
-- bytes), `bak_20260903_gripstyles_plan`, `scratchpad/grip-styles-BEFORE.html` (88,509 bytes).
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

BEGIN;

DO $$
DECLARE n int;
BEGIN
  -- the plan must still be the 11 sections stage 1 wrote
  SELECT count(*) INTO n FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
   WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name='grip-styles';
  IF n <> 11 THEN RAISE EXCEPTION 'plan is % sections, expected 11 — stage 1 has been altered, do not rebuild', n; END IF;

  -- ⚠ THIS GUARD WAS A HARD `IF n <> 5 THEN RAISE EXCEPTION` WHEN I WROTE THIS FILE 20 MINUTES
  -- AGO, AND IT WAS THE WRONG WAY ROUND. It waited for the five illustrations before allowing the
  -- rebuild. But `emit_imagery_items_action.go` is "the build-time emitter, invoked as a
  -- build-site-planner workflow step", and its own comment says status `triaged` means
  -- "build path auto-dispatch" — so a `triaged` imagery item is drained BY a build, and an
  -- operator-inserted one with no build running has nothing to dispatch it. The runbook says the
  -- same from the operations side: "Dispatch is one site at a time against a fleet-wide pool —
  -- priority 5 orders WITHIN a site, it does not jump the queue ahead of other sites."
  -- [MEASURED 2026-09-03 11:56Z] my five sat `triaged` for 16 minutes untouched, and
  -- boxingonline.com has had one waiting 2.5 HOURS — so "wait longer" is not a plan.
  -- The dependency is therefore INVERTED: the rebuild is the thing most likely to dispatch them.
  -- So this is now an observation, not a gate. If the figures are absent at build time,
  -- `Illustrated Text Block.image_url` is `required:false, on_missing:skip_field` and each section
  -- renders as plain prose — no empty frame — and the asset landing later files `image_landed`,
  -- which is one of only two re-render reasons that re-resolve. The page converges either way.
  SELECT count(*) INTO n FROM assets
   WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND status='active'
     AND asset_key IN ('illustration_ring_grip','illustration_razor_grip','illustration_shark_grip',
                       'illustration_smooth_barrel','illustration_combination_grip');
  RAISE NOTICE 'grip illustrations active at rebuild time: % of 5 (0 is expected and fine — sections build as prose and image_landed re-resolves them)', n;

  SELECT count(*) INTO n FROM site_work_items
   WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND item_type='needs_content_page'
     AND status NOT IN ('complete','cancelled','rejected','failed') AND spec->>'page_name'='grip-styles';
  IF n <> 0 THEN RAISE EXCEPTION '% open needs_content_page already target grip-styles', n; END IF;
END $$;

INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, priority,
   handler_agent, status, created_by, item_key, approval_mode)
VALUES (
  '5fe8785b-223d-41a3-88ee-c07187622381', 'operator', 'build', 'needs_content_page', 'medium',
  'Rebuild /blog/grip-styles.html into its new 11-section composition (owner 2026-08-31)',
  jsonb_build_object(
    'check',     'gap_plan_new_page',
    'page_url',  '/blog/grip-styles.html',
    'page_name', 'grip-styles',
    'reason',    'owner 2026-08-31: the guides want an accurate image per small section (ring grip, razor grip, shark grip). The page was hero + article-body + call-to-action, and article-body owns the whole article in one llm field, so a figure placed in it dies at the next body rewrite. The plan now carries 11 sections with a distinct subject each and five section-scope illustrations bound by scope_ref ordinal (IMG-075). Content rebuild through the writer, not a rerender: only the writer can redistribute one article body across nine prose sections.',
    'sections',  jsonb_build_array(
        'hero','Generic Text Block','Illustrated Text Block','Illustrated Text Block',
        'Illustrated Text Block','Illustrated Text Block','Illustrated Text Block',
        'Generic Text Block','Generic Text Block','Generic Text Block','call-to-action'),
    'suggestion',
      'Rebuild /blog/grip-styles.html as the guide it already is, now in eleven sections in this order: the hero; one text block on what grip actually changes about a throw; five illustrated blocks, one per grip style, in the order ring, razor, shark, smooth or minimal-texture, and combination; three text blocks on how grip works with barrel shape, on matching grip to your release, and on how grip needs change over time; and the closing call to action. '
      || 'Same purpose and same reader as the live page: a buying guide that helps a darts player work out which barrel texture suits their throw, written for someone choosing darts rather than for someone who already knows the vocabulary. '
      || 'EACH SECTION HAS ITS OWN SUBJECT IN THE PLAN. Write to that subject and only that subject. Do not restate the whole article in every block, do not open each block by re-introducing grip in general, and do not repeat the same comparison in more than one block — the five grip blocks are read one after another and near-identical openings are the failure mode here. '
      || 'The five illustrated blocks each carry their own photograph supplied by the framework; describe the grip in the prose and never write image, figure or iframe markup. '
      || 'Follow the site content_direction and the house voice, British English, plain human prose. State facts in the positive form. Facts only from evidence_base — the site sells no darts and takes no orders, so make no claim about stock, shipping, price or availability.'
  ),
  40, 'page-build-handler', 'triaged', 'dartsonline-grip-styles-2026-09-03b',
  'needs_content_page:grip-styles:recompose-2026-09-03', 'auto');

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_work_items WHERE created_by='dartsonline-grip-styles-2026-09-03b';
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 rebuild item, found %', n; END IF;
  -- the item's section list must match the plan EXACTLY, in order, or the binding stands down
  SELECT count(*) INTO n FROM (
    SELECT sps.ordering, sps.component_name AS planned,
           (SELECT (spec->'sections'->>sps.ordering) FROM site_work_items
             WHERE created_by='dartsonline-grip-styles-2026-09-03b') AS filed
    FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
    WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name='grip-styles') t
  WHERE planned IS DISTINCT FROM filed;
  IF n <> 0 THEN RAISE EXCEPTION '% section(s) differ between the plan and the filed item', n; END IF;
END $$;

SELECT id, status, spec->>'page_url' AS page_url, jsonb_array_length(spec->'sections') AS sections
FROM site_work_items WHERE created_by='dartsonline-grip-styles-2026-09-03b';

COMMIT;
