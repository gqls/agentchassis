-- Three card images for the hero-card-carousel going on robot-hands.com/index.html
-- at position 2, one per existing feature block.
--
-- WHY kind/purpose = 'content_hero' AND NOT 'icon'. Read from the site's live
-- imagery_style_guide (site_specs, is_current), not guessed:
--   kinds.content_hero -> medium "flat duotone editorial illustration",
--                         palette "deep charcoal ground, electric blue (#0080FF)
--                         flat shapes and linework"
--   site default       -> medium "industrial photography of real robotic grippers"
-- Only content_hero carries the blue line-art override, and that is the look of
-- the existing card-tool-*.jpg images these will sit beside. Asking for 'icon'
-- would fall through to the site default and return PHOTOGRAPHS — visually wrong
-- next to the tool cards, and the failure would look like a model problem rather
-- than a config one.
--
-- Prompts are SUBJECT-ONLY on purpose: the style guide prepends medium/mood/
-- palette at generation time (imagery_style_guide.go), which is how every other
-- robot-hands content_hero prompt in site_work_items is written. Repeating the
-- palette inline (as dartsonline's icon prompts do) would double-direct it.
--
-- scope='section', scope_ref='index:2' — the "several images for one section"
-- shape dartsonline used for its icon row. Classify() returns (100,'low') for
-- section scope, so priority is set explicitly here instead; BrandUpdate() is
-- false for section scope, which is wanted (these must not touch the brand slot).
--
-- LANDMINE AHEAD, at DEPLOY not generation (bugs_open/155, LANDMINES.md:542):
-- all three share purpose='content_hero', and deploy_image_asset_action's
-- resolveStorageURIFromAsset Priority 1 resolves the source via
-- sites.content_data->>'{purpose}_uri' — keyed on PURPOSE ONLY, never asset_id.
-- So an asset_id-only deploy of these three would fetch ONE cached URI and write
-- three byte-identical files, each reporting success:true. When deploying:
-- pass spec.s3_uri explicitly as s3://bucket/key derived from each asset's own
-- storage_path (NOT assets.url — bugs_open/152: url is overwritten post-deploy),
-- then sha256sum all three and confirm they differ before believing any of it.

\set site '00ff3af5-dad8-4770-9f70-3edc267a3c92'

BEGIN;

INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, status, created_by, pipeline,
   priority, triaged_at, handler_agent, item_key, spec)
SELECT
  :'site'::uuid, 'robot-hands-carousel', 'needs_imagery', 'medium',
  'Carousel card image: ' || c.label,
  'triaged', 'session-2026-07-31-robot-hands-carousel', 'build',
  60, now(), 'image-build-handler',
  'needs_imagery:section:index:2:' || c.asset_key,
  jsonb_build_object(
    'key',           c.asset_key,
    'asset_key',     c.asset_key,
    'kind',          'content_hero',
    'purpose',       'content_hero',
    'scope',         'section',
    'scope_ref',     'index:2',
    'check',         'robot-hands-carousel',
    'prompt',        c.prompt,
    'style_hints',   jsonb_build_object('aspect_ratio','16:9'),
    'brand_update',  false
  )
FROM (VALUES
  ('content_hero_rh_carousel_catalog',
   'Cross-Technology Catalog',
   'Four distinct robotic gripper end-effectors arranged side by side for comparison — a two-finger parallel-jaw gripper, a vacuum suction-cup array, a magnetic pad, and a soft bellows gripper — bold simple silhouettes of equal weight, no one favoured. No text, lettering, numbers or labels.'),
  ('content_hero_rh_carousel_matchmatrix',
   'MatchMatrix Scoring Engine',
   'A ranked comparison matrix beside a robotic gripper silhouette: a column of candidate rows with one row clearly marked as the best match by a bold tick, the rest plainly ranked below it. A selection diagram, not a chart of data. No text, lettering, numbers or labels.'),
  ('content_hero_rh_carousel_calculators',
   'Payload & Parameter Calculators',
   'A two-finger robotic gripper holding a cube, surrounded by engineering annotation: force arrows pressing inward on the jaws, a downward weight arrow under the cube, and dimension lines measuring the jaw opening. A statics free-body diagram. No text, lettering, numbers or labels.')
) AS c(asset_key, label, prompt)
-- Conflict target must match idx_swi_dedup EXACTLY, predicate included, or
-- Postgres refuses to infer the index. Terminal-status list copied verbatim from
-- the live index definition (= workItemTerminalStatuses in work_items_common.go;
-- these two are one contract and drift shows up as a fleet-wide 42P10).
ON CONFLICT (site_id, item_key) WHERE item_key IS NOT NULL AND (status <> ALL (ARRAY[
  'complete'::text, 'verified'::text, 'rejected'::text, 'wont_fix'::text,
  'failed'::text, 'unresolved'::text, 'cancelled'::text]))
DO NOTHING;

\echo '--- queued imagery ---'
SELECT summary, status, priority, spec->>'asset_key' AS asset_key,
       spec->>'kind' AS kind, spec->'style_hints'->>'aspect_ratio' AS aspect
FROM site_work_items
WHERE site_id = :'site' AND source = 'robot-hands-carousel'
ORDER BY spec->>'asset_key';

COMMIT;
