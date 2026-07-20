-- R4c (2026-07-20) — the remaining two fabricated stat blocks.
--
-- R4b fixed `about`. The verify query in that file surfaced a second block, and
-- a site-wide sweep for `stat[_]?N_?value` found a third. All three are now
-- accounted for:
--
--   about / content-block-about   FIXED in R4b  (was 1,200+ / 6 / "Tracked & Published")
--   gripper-detail / system-stats THIS FILE     (2,400+ / 140+ / 6 / 18)  — LIVE, 200
--   index / brief-explanation     THIS FILE     (— / 6 / —)               — LIVE, 200
--
-- The gripper-detail block is the worst of the three. Beyond the values, its
-- suffixes are unedited generic-template placeholders — "%" on gripper models,
-- "ms" on manufacturers — and its descriptions assert things the site does not
-- do: "datasheet sources and last-verified dates are displayed at model level"
-- (they are not) and scoring "across 18 parameters including ... cycle time (ms),
-- payload-to-weight ratio, and operating pressure range" (none of which exist in
-- `products.specifications`).
--
-- Note `index / brief-explanation` had already had two of its three values
-- blanked to an em dash — someone recognised they had no source — but left the
-- fabricated "6 actuation technologies" in place. A partial cleanup reads as a
-- checked block, which is worse than an obviously empty one.
--
-- GROUND TRUTH (queried 2026-07-20, this is where every figure below comes from):
--   SELECT count(*)                                        -> 5   gripper models
--   SELECT count(DISTINCT specifications->>'manufacturer') -> 5   manufacturers
--   SELECT count(*) FROM products, LATERAL jsonb_each_text(specifications)
--          WHERE key <> 'manufacturer'                     -> 24  published figures
--   ...all WHERE site_id='00ff3af5-...' AND category='gripper'
--
-- "Parameters compared = 4" is payload, jaw travel, gripping force, IP rating —
-- exactly the four MatchMatrix tests against. There is NO actuation-type field
-- in the data, so no honest actuation-technology count exists at any value.
--
-- Apply:  kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--           psql -U clients_user -d clients_db < THIS_FILE

\set site '00ff3af5-dad8-4770-9f70-3edc267a3c92'

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. gripper-detail / system-stats — all four values, suffixes and descriptions.
-- ---------------------------------------------------------------------------
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('stat1_label', 'Gripper Models Indexed')
    || jsonb_build_object('stat1_value', '5')
    || jsonb_build_object('stat1_suffix', '')
    || jsonb_build_object('stat1_description',
         'Every gripper in the index carries the figures its manufacturer '
      || 'publishes, normalised so they compare like for like. The index is small '
      || 'and fully sourced rather than broad and estimated.')

    || jsonb_build_object('stat2_label', 'Manufacturers Covered')
    || jsonb_build_object('stat2_value', '5')
    || jsonb_build_object('stat2_suffix', '')
    || jsonb_build_object('stat2_description',
         'Schunk, OnRobot, Robotiq, Zimmer Group and Festo. Figures are '
      || 'reproduced as each manufacturer publishes them; where one does not '
      || 'publish a parameter it is shown as unpublished, never estimated.')

    || jsonb_build_object('stat3_label', 'Parameters Compared')
    || jsonb_build_object('stat3_value', '4')
    || jsonb_build_object('stat3_suffix', '')
    || jsonb_build_object('stat3_description',
         'Payload, jaw travel, gripping force and IP rating — the four parameters '
      || 'MatchMatrix tests an application against. Each result shows which '
      || 'criterion passed, which failed, and which the manufacturer leaves blank.')

    || jsonb_build_object('stat4_label', 'Published Figures Held')
    || jsonb_build_object('stat4_value', '24')
    || jsonb_build_object('stat4_suffix', '')
    || jsonb_build_object('stat4_description',
         'Individual specification figures across the indexed grippers. Coverage '
      || 'is uneven by design: two models publish no payload rating and two '
      || 'publish no IP rating, and MatchMatrix flags those gaps rather than '
      || 'filling them.'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND pc.content_data->>'stat1_value' = '2,400+';

-- ---------------------------------------------------------------------------
-- 2. index / brief-explanation — fill the two em dashes, drop the "6".
-- ---------------------------------------------------------------------------
UPDATE page_components pc SET
  content_data = pc.content_data
    || jsonb_build_object('stat_1_label', 'Gripper models indexed')
    || jsonb_build_object('stat_1_value', '5')
    || jsonb_build_object('stat_2_label', 'Parameters compared')
    || jsonb_build_object('stat_2_value', '4')
    || jsonb_build_object('stat_3_label', 'Manufacturers covered')
    || jsonb_build_object('stat_3_value', '5'),
  updated_at = now()
FROM pages p
WHERE p.id = pc.page_id AND p.site_id = :'site'
  AND p.name = 'index'
  AND pc.content_data->>'stat_2_label' = 'Actuation technologies benchmarked';

-- ---------------------------------------------------------------------------
-- VERIFY
-- ---------------------------------------------------------------------------
\echo '--- every stat value on the site, with its label (all must be sourced) ---'
SELECT p.name, e.k, e.v
FROM page_components pc JOIN pages p ON p.id = pc.page_id,
LATERAL jsonb_each_text(pc.content_data) AS e(k,v)
WHERE p.site_id = :'site' AND e.k ~ 'stat[_]?[0-9]+_?(value|label)'
ORDER BY p.name, e.k;

\echo '--- any surviving six-technology claim anywhere in content_data (expect 0) ---'
SELECT p.name, e.k
FROM page_components pc JOIN pages p ON p.id = pc.page_id,
LATERAL jsonb_each_text(pc.content_data) AS e(k,v)
WHERE p.site_id = :'site'
  AND v ILIKE '%soft-robotic%'
ORDER BY p.name;

COMMIT;
