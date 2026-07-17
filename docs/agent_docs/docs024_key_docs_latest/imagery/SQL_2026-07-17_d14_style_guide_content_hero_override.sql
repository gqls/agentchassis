-- D14 (2026-07-17, I3 gate fix): per-kind style-guide override for the new
-- content_hero kind — flat duotone editorial illustration, replacing the
-- photographic SDXL voice that failed the D13 card gate (colour drift,
-- medium drift, text artefacts). Supersede-row pattern, not in-place.
--
-- Inert until the v1.0.1132+ binary (commit 4e35c8064) is live: the old
-- binary's imageryStyleGuide struct has no `kinds` field, so json.Unmarshal
-- ignores it; guide-level fields are unchanged, so pre-rollout behaviour is
-- identical. The override REPLACES guide-level fields wholesale for
-- content_hero — including the empty reference_asset_keys (no photographic
-- anchors for a flat-illustration kind) and the avoid list (the base avoid
-- forbids "cartoonish rendering", which would fight this style).
--
-- Site: robot-hands.com 00ff3af5-dad8-4770-9f70-3edc267a3c92
-- Supersedes: 439329c4-b705-48bd-81af-5a86a9b8d7db (seeded 2026-07-10, I1)

BEGIN;

UPDATE site_specs
   SET is_current = false,
       superseded_at = now(),
       updated_at = now()
 WHERE id = '439329c4-b705-48bd-81af-5a86a9b8d7db'
   AND is_current = true;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT site_id,
       aspect,
       data || jsonb_build_object('kinds', jsonb_build_object('content_hero', jsonb_build_object(
           'medium',  'flat duotone editorial illustration',
           'mood',    'bold simple silhouette of the subject, minimal detail, precise, technical',
           'palette', 'deep charcoal ground, electric blue (#0080FF) flat shapes and linework, light grey secondary accents only',
           'avoid',   'photorealism, photographic texture, gradients, 3D rendering, drop shadows, text, lettering, logos, watermarks, busy detail, colour outside the palette',
           'reference_asset_keys', jsonb_build_array()
       ))),
       source,
       source_agent,
       'D14 2026-07-17: content_hero per-kind override — flat duotone editorial illustration (I3 card-gate fix). Supersedes 439329c4-b705-48bd-81af-5a86a9b8d7db.',
       true,
       'imagery-i3-d14'
  FROM site_specs
 WHERE id = '439329c4-b705-48bd-81af-5a86a9b8d7db';

COMMIT;
