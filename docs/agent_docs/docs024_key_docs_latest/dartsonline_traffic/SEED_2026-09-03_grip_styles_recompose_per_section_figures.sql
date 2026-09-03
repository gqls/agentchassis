-- SEED_2026-09-03_grip_styles_recompose_per_section_figures.sql
--
-- STAGE 1 of the grip-styles recompose: the PLAN, the SUBJECTS, the imagery rows, and one
-- style-guide gap. Applies no content and builds nothing — stage 2 generates the five images,
-- stage 3 rebuilds the page through the writer.
--
-- WHY. Owner, 2026-08-31: the guides want an accurate image per small section (ring grip, razor
-- grip, shark grip…). `/blog/grip-styles.html` is the owner's own example. It is today three
-- components — `hero`, `article-body`, `call-to-action` — and `article-body`'s single `content`
-- field is documented "Write the full article body as HTML", so prose and any figure share one
-- llm-owned field and a figure spliced in dies at the next body rewrite. That is §3.2's
-- "one field owns prose and figures" in its purest form.
--
-- WHAT MAKES THIS POSSIBLE NOW: `IMG-075` (live 2026-09-01 r1 / 2026-09-02 15:39 r2, chassis
-- v1.0.1354) binds a section-scope `site_plan_imagery` row to ONE section via its `scope_ref`
-- ordinal, so a page can carry a different figure per section and each re-derives from the plan
-- on every build and re-render. `Illustrated Text Block` is one of exactly TWO components in the
-- estate whose `image_url` is sourced `site_assets.illustration` (the other is
-- `brief-explanation`) [MEASURED 2026-09-02]; its `image_url`/`image_alt` are
-- `required:false, on_missing:skip_field`, so a section whose figure has not landed renders as
-- plain prose rather than an empty frame.
--
-- ⚠ ORDINAL SEMANTICS — READ THE CODE, NOT THE NEIGHBOURING SITE. `sectionRefForOrdinal`
-- (plan_sections_action.go) treats the ordinal as a **0-based index into `planSectionOrder`**,
-- and `planSectionOrder` returns EVERY `site_plan_sections` row for the page ordered by
-- `ordering`, explicitly "including any site-level slots" — it filters nothing. So the ordinal
-- indexes THIS FILE'S `ordering` column directly. `sectionScopeRefOrdinal` takes everything after
-- the LAST colon, so `grip-styles:2` → 2. A figure bound one section out "renders and deploys
-- looking exactly right", which is why this is spelled out rather than copied from apis.uk.
--
-- ⚠ SUBJECTS ARE NOT OPTIONAL HERE AND NOTHING WILL STOP ME OMITTING THEM. This page is about to
-- carry 5 repeats of `Illustrated Text Block` and 4 of `Generic Text Block`. The detector for
-- exactly that (`REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT`, plan_sections_action.go, filed for
-- `bugs_open/443`) is **observe-only, severity "warning"** — it logs and the build proceeds, and
-- the predicted result is "identical briefs, near-duplicate output". Its own remedy names the fix:
-- "site_plan_sections.subject for planned sites". Every section below therefore carries a distinct
-- subject, and they are written to be distinguishable to a WRITER, not merely different strings.
--
-- ⚠ THE STYLE-GUIDE GAP THIS ALSO CLOSES, found while writing the imagery rows. On 2026-08-31 I
-- appended darts-anatomy prohibitions to `kinds.hero.avoid` and `kinds.content_hero.avoid`,
-- because `avoidForKind` returns a per-kind override INSTEAD of the guide-level list. These five
-- figures are `kind='illustration'`, which has NO override — so `avoidForKind('illustration')`
-- returns the GUIDE-LEVEL avoid, which is the one place I did not put the anatomy clauses. Close-up
-- grip photography is precisely where barrel anatomy matters. So the clauses go to guide level too;
-- the duplication across the two overrides is harmless because only one string is ever used per
-- kind, and guide level now covers every kind that has no override.
--
-- RESTORE POINTS (taken 2026-09-03 before this file):
--   bak_20260903_gripstyles_components  — 3 rows, article-body content_data 8,401 bytes
--   bak_20260903_gripstyles_plan        — the 3 plan rows
--   scratchpad/grip-styles-BEFORE.html  — 88,509 bytes as served
--
-- Site: dartsonline.com  5fe8785b-223d-41a3-88ee-c07187622381
-- Plan: 0fb05b75-04f4-4f4c-8890-c34d6a71012c (current)

BEGIN;

-- ── 0. pre-state assertions ──────────────────────────────────────────────────
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM bak_20260903_gripstyles_plan;
  IF n <> 3 THEN RAISE EXCEPTION 'plan backup missing or wrong size (%) — do not proceed without a restore point', n; END IF;

  SELECT count(*) INTO n FROM bak_20260903_gripstyles_components;
  IF n <> 3 THEN RAISE EXCEPTION 'component backup missing or wrong size (%)', n; END IF;

  -- both components must exist and be active, or the plan names something unbuildable
  SELECT count(*) INTO n FROM content_components
   WHERE name IN ('Illustrated Text Block','Generic Text Block','hero','call-to-action') AND is_active;
  IF n <> 4 THEN RAISE EXCEPTION 'expected 4 active components (Illustrated Text Block, Generic Text Block, hero, call-to-action), found %', n; END IF;

  -- the illustration source must still be declared, or the figures can never resolve
  IF (SELECT input_schema->'fields'->'image_url'->>'source' FROM content_components
        WHERE name='Illustrated Text Block' AND is_active) <> 'site_assets.illustration' THEN
    RAISE EXCEPTION 'Illustrated Text Block no longer sources site_assets.illustration — the whole binding premise is gone';
  END IF;

  SELECT count(*) INTO n FROM site_plan_imagery spi
    JOIN site_plans sp ON sp.id=spi.plan_id AND sp.is_current
   WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND spi.scope='section' AND spi.scope_ref LIKE 'grip-styles:%';
  IF n <> 0 THEN RAISE EXCEPTION '% section-scope imagery row(s) already exist for grip-styles', n; END IF;
END $$;

-- ── 1. the composition ───────────────────────────────────────────────────────
-- Ordering is the ordinal. Keep this table and the imagery scope_refs below in step.
--   0 hero                     6 Illustrated  combination grips      <- figure
--   1 Generic      what grip changes          7 Generic  barrel shape
--   2 Illustrated  ring grip       <- figure  8 Generic  release
--   3 Illustrated  razor grip      <- figure  9 Generic  hands change
--   4 Illustrated  shark grip      <- figure 10 call-to-action
--   5 Illustrated  smooth barrels  <- figure
DELETE FROM site_plan_sections sps
USING site_plans sp
WHERE sp.id = sps.plan_id AND sp.is_current
  AND sp.site_id = '5fe8785b-223d-41a3-88ee-c07187622381'
  AND sps.page_name = 'grip-styles';

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

-- ── 2. the five figures, bound to sections 2–6 ───────────────────────────────
-- kind='illustration' so the per-section map's kind arm can reach them and
-- Illustrated Text Block's `site_assets.illustration` source resolves.
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

-- ── 3. the style-guide gap: illustration has no override, so guide level is what it reads ──
CREATE TEMP TABLE _guide ON COMMIT DROP AS
SELECT jsonb_set(data, '{avoid}',
         to_jsonb((data #>> '{avoid}') ||
           '; feathered or fletched flights of any kind — archery fletching on a dart; an all-red bull with no green outer ring; dartboard numbers out of their true sequence, repeated, invented or garbled; board segments coloured blue, purple, orange or anything other than black and cream; concentric rings in place of radial segments; darts drawn as arrows; a barrel rendered as a screw thread, drill bit or bolt')) AS data
FROM site_specs
WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND aspect='imagery_style_guide' AND is_current;

UPDATE site_specs SET is_current=false, superseded_at=now()
WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND aspect='imagery_style_guide' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, notes)
SELECT '5fe8785b-223d-41a3-88ee-c07187622381', 'imagery_style_guide', data,
  'hand_authored', 'dartsonline-traffic-2026-09-03', 'dartsonline-traffic-2026-09-03',
  'Extends the darts-anatomy prohibitions to the GUIDE-LEVEL avoid. On 2026-08-31 they went only '
  || 'into kinds.hero and kinds.content_hero, because avoidForKind returns an override INSTEAD of '
  || 'the guide-level list. kind=illustration has no override, so it reads guide level — the one '
  || 'place the clauses were missing, and the kind the grip figures use. Adds one clause the '
  || 'close-up barrel work needs: no barrel rendered as a screw thread, drill bit or bolt.'
FROM _guide;

-- ── 4. post-state assertions ─────────────────────────────────────────────────
DO $$
DECLARE n int; g jsonb; ord text;
BEGIN
  SELECT count(*) INTO n FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
   WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name='grip-styles';
  IF n <> 11 THEN RAISE EXCEPTION 'expected 11 plan sections, found %', n; END IF;

  -- every repeated component must carry a subject, or 443's symptom is reproduced here
  SELECT count(*) INTO n FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
   WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name='grip-styles'
     AND sps.component_name IN ('Illustrated Text Block','Generic Text Block')
     AND (sps.subject IS NULL OR length(trim(sps.subject)) < 40);
  IF n <> 0 THEN RAISE EXCEPTION '% repeated-component section(s) carry no usable subject', n; END IF;

  -- subjects must be DISTINCT, not merely present
  SELECT count(*) INTO n FROM (
    SELECT sps.subject FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
     WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name='grip-styles' AND sps.subject IS NOT NULL
     GROUP BY 1 HAVING count(*) > 1) d;
  IF n <> 0 THEN RAISE EXCEPTION '% duplicated subject(s)', n; END IF;

  -- ordinals must be in range AND land on Illustrated Text Block, or figures bind to prose
  FOR ord IN
    SELECT spi.scope_ref FROM site_plan_imagery spi JOIN site_plans sp ON sp.id=spi.plan_id AND sp.is_current
     WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND spi.scope='section' AND spi.scope_ref LIKE 'grip-styles:%'
  LOOP
    SELECT count(*) INTO n FROM site_plan_sections sps JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
     WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name='grip-styles'
       AND sps.ordering = split_part(ord, ':', 2)::int
       AND sps.component_name = 'Illustrated Text Block';
    IF n <> 1 THEN RAISE EXCEPTION 'scope_ref % does not index an Illustrated Text Block section', ord; END IF;
  END LOOP;

  SELECT count(*) INTO n FROM site_plan_imagery spi JOIN site_plans sp ON sp.id=spi.plan_id AND sp.is_current
   WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND spi.scope='section' AND spi.scope_ref LIKE 'grip-styles:%';
  IF n <> 5 THEN RAISE EXCEPTION 'expected 5 section figures, found %', n; END IF;

  SELECT data INTO g FROM site_specs WHERE site_id='5fe8785b-223d-41a3-88ee-c07187622381'
    AND aspect='imagery_style_guide' AND is_current;
  IF (g #>> '{avoid}') NOT LIKE '%fletched flights%' THEN RAISE EXCEPTION 'guide-level anatomy clause did not land'; END IF;
  IF (g #>> '{avoid}') NOT LIKE '%identifiable faces%' THEN RAISE EXCEPTION 'pre-existing guide-level avoid lost'; END IF;
  IF (g #>> '{kinds,hero,avoid}') NOT LIKE '%catalogue shot%' THEN RAISE EXCEPTION 'hero override damaged'; END IF;
END $$;

SELECT sps.ordering, sps.component_name, COALESCE(spi.key,'—') AS figure, left(COALESCE(sps.subject,''),52) AS subject
FROM site_plan_sections sps
JOIN site_plans sp ON sp.id=sps.plan_id AND sp.is_current
LEFT JOIN site_plan_imagery spi ON spi.plan_id=sp.id AND spi.scope='section'
     AND split_part(spi.scope_ref,':',2)::int = sps.ordering
WHERE sp.site_id='5fe8785b-223d-41a3-88ee-c07187622381' AND sps.page_name='grip-styles'
ORDER BY sps.ordering;

COMMIT;
