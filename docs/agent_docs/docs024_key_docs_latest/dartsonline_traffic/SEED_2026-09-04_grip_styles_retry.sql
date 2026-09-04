-- SEED_2026-09-04_grip_styles_retry.sql
--
-- THE RETRY. Re-applies the 11-section plan, the nine subjects and the five section-bound figures,
-- then files the rebuild — now that the blocker is cleared.
--
-- WHY NOW. Both halves of the subject fix are live, established at the ARTEFACT because the
-- migrations ledger could not answer it in either direction:
--   * writer receives the subject — `641` applied BY HAND 2026-09-03 22:05:57Z, NO row in
--     `schema_migrations`. Live template: `{{if .current_section.subject}}## This section
--     {{.current_section.subject}}` plus a sibling list of every OTHER section's subject.
--   * planner emits subjects — `640` never ran and is SUPERSEDED by `762`, applied AND recorded
--     2026-09-03 19:22:35Z. 640's own anchor `may also carry a "subject"` returns f.
--
-- WHAT IS DELIBERATELY NOT REPEATED FROM `SEED_2026-09-03`:
--   * §3, the style-guide anatomy clause. It survived the revert (the revert touched
--     page_components / site_plan_sections / site_plan_imagery / pages.sections, never
--     site_specs). [MEASURED 2026-09-04] the guide-level avoid already contains
--     'fletched flights' exactly ONCE — re-running §3 would append it a second time.
--   * stage 2, image generation. All five illustrations are already active assets.
--
-- The plan and imagery blocks below are LIFTED VERBATIM from SEED_2026-09-03 (the authority),
-- not retyped — the subjects and prompts are byte-identical to the ones already reviewed.
--
-- Chassis verified on the post-roll build before dispatch: both LIVE pods (matched by name against
-- `kubectl get pods`, not by capability row alone) stamped 06c0b18f2, up 16:01:30Z / 16:02:05Z,
-- `rollout status` complete, >300s elapsed.
--
-- HOW TO GRADE IT — and the test we previously agreed is RETIRED. "N sections -> N distinct prompt
-- hashes" CANNOT FAIL post-641: the sibling block lists every subject except the current one, so
-- prompts differ structurally even if the `## This section` block were empty. Grade on the
-- PER-PROMPT `## This section` block, and on the served prose.
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
   WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name='grip-styles';
  IF n <> 3 THEN RAISE EXCEPTION 'plan is % sections, expected the reverted 3 — state is not what this file assumes', n; END IF;

  SELECT count(*) INTO n FROM site_plan_imagery spi JOIN site_plans sp ON sp.id=spi.plan_id AND sp.is_current
   WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND spi.scope='section' AND spi.scope_ref LIKE 'grip-styles:%';
  IF n <> 0 THEN RAISE EXCEPTION '% section imagery row(s) already present', n; END IF;

  -- the five figures must still exist, or this becomes stage 2 as well
  SELECT count(*) INTO n FROM assets WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND status='active'
     AND asset_key IN ('illustration_ring_grip','illustration_razor_grip','illustration_shark_grip',
                       'illustration_smooth_barrel','illustration_combination_grip');
  IF n <> 5 THEN RAISE EXCEPTION 'only % of 5 grip illustrations are active — regenerate before retrying', n; END IF;

  -- the blocker must actually be cleared; escaped underscore, and the control must fire
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config::text LIKE '%{{.current\_section.subject}}%';
  IF n <> 1 THEN RAISE EXCEPTION 'writer does not interpolate current_section.subject (found %) — 641 is not live, do not retry', n; END IF;

  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
     AND default_config::text LIKE '%current\_section.subjectNOTREAL%';
  IF n <> 0 THEN RAISE EXCEPTION 'must-be-absent control matched — the gate is not discriminating'; END IF;

  -- do not re-append the style-guide clause
  SELECT (length(data #>> '{avoid}') - length(replace(data #>> '{avoid}','fletched flights','')))/16 INTO n
    FROM site_specs WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND aspect='imagery_style_guide' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'anatomy clause occurs % times in the guide-level avoid, expected exactly 1', n; END IF;
END $$;

DELETE FROM site_plan_sections sps USING site_plans sp
WHERE sp.id = sps.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name = 'grip-styles';

INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name, subject)
SELECT sp.id, 'grip-styles', v.ordering, v.component_name, v.subject
FROM site_plans sp,
(VALUES
  (0,  'hero',                   NULL),
  (1,  'Generic Text Block',     'What grip actually changes about a throw: how barrel texture governs pressure, release point and repeatability, and why choosing a grip is a question of fit rather than of technique'),
  (2,  'Illustrated Text Block', 'Ring grip: evenly spaced circular grooves cut around the barrel, the most common and most forgiving texture, who it suits and how it feels on release'),
  (3,  'Illustrated Text Block', 'Razor grip: sharp, deep, closely spaced machined edges, the most aggressive texture available, what it gives a thrower who needs certainty and when it is too much'),
  (4,  'Illustrated Text Block', 'Shark grip: angled directional cuts that hold the fingers during the throw and release cleanly, and why the direction of the cut is the whole point'),
  (5,  'Illustrated Text Block', 'Smooth and minimal-texture barrels: why some throwers want almost no grip at all, what a polished or lightly milled barrel does for a fast release, and who should avoid one'),
  (6,  'Illustrated Text Block', 'Combination grips: distinct textured zones along one barrel, how to read where each zone is meant to sit under the fingers, and why they suit a thrower still settling their hold'),
  (7,  'Generic Text Block',     'How grip and barrel shape work together: texture is only half of what the hand feels, and the same grip pattern behaves differently on a straight, torpedo or bomb barrel'),
  (8,  'Generic Text Block',     'Matching grip to your release: reading whether you have a fast clean release or a slower heavier one, and which textures suit each'),
  (9,  'Generic Text Block',     'How grip needs change over time: hand temperature, moisture, skin condition and barrel wear, and why the grip that suited you last year may not suit you now'),
  (10, 'call-to-action',         NULL)
) AS v(ordering, component_name, subject)
WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND sp.is_current;

INSERT INTO site_plan_imagery (plan_id, scope, scope_ref, key, kind, prompt, style_hints, ordering, source)
SELECT sp.id, 'section', v.scope_ref, v.key, 'illustration', v.prompt,
       jsonb_build_object('aspect_ratio','4:3'), v.ordering, 'manual'
FROM site_plans sp,
(VALUES
  ('grip-styles:2', 'illustration_ring_grip',
   'Extreme close-up of a single tungsten dart barrel showing RING grip: evenly spaced circular grooves machined right around the barrel, perpendicular to its long axis, each groove a clean parallel band with smooth metal between. Hard directional rim light picking out the machined edges, near-black unlit background, shallow depth of field. Physical accuracy matters: this is a darts barrel, not a screw thread or a drill bit — it tapers, it is knurled steel-grey tungsten, and any visible end carries a fine steel point or a slim shaft, never a feather.', 0),
  ('grip-styles:3', 'illustration_razor_grip',
   'Extreme close-up of a single tungsten dart barrel showing RAZOR grip: sharp, deep, tightly spaced machined cuts with knife-like crests between them, visibly more aggressive than a plain ring pattern. Hard directional rim light catching the sharp edges so the texture reads as sharp, near-black unlit background, shallow depth of field. Physical accuracy matters: this is a darts barrel, tapering knurled steel-grey tungsten; any visible end carries a fine steel point or a slim shaft, never a feather.', 1),
  ('grip-styles:4', 'illustration_shark_grip',
   'Extreme close-up of a single tungsten dart barrel showing SHARK grip: angled, directional cuts raked consistently one way along the barrel like fish scales, so the texture visibly bites in one direction and releases in the other. Hard directional rim light raking across the angled cuts, near-black unlit background, shallow depth of field. Physical accuracy matters: this is a darts barrel, tapering knurled steel-grey tungsten; any visible end carries a fine steel point or a slim shaft, never a feather.', 2),
  ('grip-styles:5', 'illustration_smooth_barrel',
   'Extreme close-up of a single tungsten dart barrel that is SMOOTH: polished or very lightly milled, almost no machined texture, a clean continuous metal surface with one soft highlight running its length. Hard directional rim light, near-black unlit background, shallow depth of field. Physical accuracy matters: this is a darts barrel, tapering steel-grey tungsten; any visible end carries a fine steel point or a slim shaft, never a feather.', 3),
  ('grip-styles:6', 'illustration_combination_grip',
   'Extreme close-up of a single tungsten dart barrel showing a COMBINATION grip: two or three clearly distinct texture zones along one barrel — for example fine rings toward the point, a smooth polished band in the middle, and deeper cuts toward the rear — with the boundaries between zones plainly visible. Hard directional rim light, near-black unlit background, shallow depth of field. Physical accuracy matters: this is a darts barrel, tapering knurled steel-grey tungsten; any visible end carries a fine steel point or a slim shaft, never a feather.', 4)
) AS v(scope_ref, key, prompt, ordering)
WHERE sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381' AND sp.is_current;

INSERT INTO site_work_items
  (site_id, source, pipeline, item_type, severity, summary, spec, priority,
   handler_agent, status, created_by, item_key, approval_mode)
VALUES (
  '5fe8785b-223d-41a3-88ee-c07187622381', 'operator', 'build', 'needs_content_page', 'medium',
  'Rebuild /blog/grip-styles.html into its 11-section composition — RETRY, subject fix now live',
  jsonb_build_object(
    'check','gap_plan_new_page','page_url','/blog/grip-styles.html','page_name','grip-styles',
    'reason','owner 2026-08-31: an accurate image per small section. First attempt 2026-09-03 built all 11 sections and bound five distinct figures correctly, but every section received a byte-identical prompt and rewrote the whole article (bugs_open/443). Migration 641 (writer receives the section subject) went live by hand 2026-09-03 22:05Z and 762 (planner emits subjects) 19:22Z, so the writer now sees this section subject and the siblings. Retry.',
    'sections', jsonb_build_array(
        'hero','Generic Text Block','Illustrated Text Block','Illustrated Text Block',
        'Illustrated Text Block','Illustrated Text Block','Illustrated Text Block',
        'Generic Text Block','Generic Text Block','Generic Text Block','call-to-action'),
    'suggestion',
      'Rebuild /blog/grip-styles.html as the guide it already is, now in eleven sections in this order: the hero; one text block on what grip actually changes about a throw; five illustrated blocks, one per grip style, in the order ring, razor, shark, smooth or minimal-texture, and combination; three text blocks on how grip works with barrel shape, on matching grip to your release, and on how grip needs change over time; and the closing call to action. '
      || 'Same purpose and same reader as the live page: a buying guide that helps a darts player work out which barrel texture suits their throw, written for someone choosing darts rather than for someone who already knows the vocabulary. '
      || 'EACH SECTION HAS ITS OWN SUBJECT. Write to that subject and only that subject. Do not restate the whole article in every block, do not open each block by re-introducing grip in general, and do not repeat the same comparison in more than one block. '
      || 'The five illustrated blocks each carry their own photograph supplied by the framework; describe the grip in the prose and never write image, figure or iframe markup. '
      || 'Follow the site content_direction and the house voice, British English, plain human prose. State facts in the positive form. Facts only from evidence_base — the site sells no darts and takes no orders, so make no claim about stock, shipping, price or availability.'
  ),
  40, 'page-build-handler', 'triaged', 'dartsonline-grip-styles-retry-2026-09-04',
  'needs_content_page:grip-styles:retry-2026-09-04', 'auto');

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
   WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name='grip-styles';
  IF n <> 11 THEN RAISE EXCEPTION 'expected 11 plan sections, found %', n; END IF;
  SELECT count(*) INTO n FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
   WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name='grip-styles'
     AND sps.component_name IN ('Illustrated Text Block','Generic Text Block')
     AND (sps.subject IS NULL OR length(trim(sps.subject)) < 40);
  IF n <> 0 THEN RAISE EXCEPTION '% repeated-component section(s) carry no usable subject', n; END IF;
  SELECT count(*) INTO n FROM site_plan_imagery spi JOIN site_plans sp ON sp.id=spi.plan_id AND sp.is_current
   WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND spi.scope='section' AND spi.scope_ref LIKE 'grip-styles:%';
  IF n <> 5 THEN RAISE EXCEPTION 'expected 5 section figures, found %', n; END IF;
  SELECT count(*) INTO n FROM site_work_items WHERE created_by='dartsonline-grip-styles-retry-2026-09-04';
  IF n <> 1 THEN RAISE EXCEPTION 'expected 1 rebuild item, found %', n; END IF;
END $$;

SELECT sps.ordering, sps.component_name, COALESCE(spi.key,'—') AS figure
FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
LEFT JOIN site_plan_imagery spi ON spi.plan_id=sp.id AND spi.scope='section' AND spi.scope_ref='grip-styles:'||sps.ordering
WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name='grip-styles' ORDER BY sps.ordering;

COMMIT;
