-- SEED — loancalculator.co.uk: plan_includes_tools:"true" into the structure
-- spec aspect (planner half 2, PLAN_2026-08-12_planner_sees_locked_tools.md;
-- pairs with migration 407, which is inert until a site carries this key).
--
-- Read by the widened load_components query in the build-site-planner
-- definition (407): with this key AND a tool already placed on this site's own
-- pages, the planner's component menu includes that tool, so a plan can name
-- the site's calculators and the identity arm (f4820a877) can pair them with
-- their locked rows instead of exiling them to the page foot.
--
-- Fifth opt-in key on this aspect (url_shape, honour_realised_identity,
-- twin_identity_snap, stem_twin_snap, now plan_includes_tools) — same idiom:
-- unsafe default OFF, absent key = today's behaviour.
--
-- Supersede-then-insert, per convention. The new row carries the OLD data
-- (pages list 27 entries, source, adopted_from, url_shape) plus the one new
-- key.
--
-- ⚠ Same standing cautions as the url_shape seed: do NOT rely on `pinned`;
--   a re-adoption WOULD have dropped opt-in keys until the carry-forward fix
--   (19acfc895, LIVE v1.0.1294+) — that fix now preserves them, but check
--   after any adoption run anyway:
--     SELECT data ?& array['url_shape','plan_includes_tools'] FROM site_specs
--     WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
--       AND aspect='structure' AND is_current;

BEGIN;

UPDATE site_specs SET is_current = false, superseded_at = now()
WHERE site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
  AND aspect = 'structure' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, notes, is_current, created_by)
SELECT
  site_id,
  'structure',
  data || '{"plan_includes_tools": "true"}'::jsonb,
  'operator',
  'plan_includes_tools seeded 2026-08-14 (planner half 2, migration 407): the planner may offer this site''s own already-placed tool components in its menu, so a replan names the 12 locked calculators instead of omitting them. Carries forward pages/source/adopted_from/url_shape unchanged.',
  true,
  'loancalculator_rebuild_thread'
FROM site_specs
WHERE site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
  AND aspect = 'structure'
ORDER BY created_at DESC
LIMIT 1;

-- Verify or abort: exactly one current row, BOTH opt-in keys present, pages
-- list intact at 27 (the adoption-written list incl. archived standard-calc —
-- the gap-planner's guide-loan-faqs page of 2026-08-14 lives in pages, not in
-- this list, so the count is unchanged from the url_shape seed's).
DO $$
DECLARE n int; has_new boolean; has_old boolean; n_pages int; v text;
BEGIN
  SELECT count(*) INTO n FROM site_specs
  WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='structure' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 current structure spec, found %', n; END IF;

  SELECT data ? 'plan_includes_tools', data ? 'url_shape',
         jsonb_array_length(data->'pages'), data->>'plan_includes_tools'
    INTO has_new, has_old, n_pages, v
  FROM site_specs
  WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='structure' AND is_current;
  IF NOT has_new THEN RAISE EXCEPTION 'plan_includes_tools key absent after seed'; END IF;
  IF v <> 'true' THEN RAISE EXCEPTION 'plan_includes_tools is %, not the string true the 407 query matches', v; END IF;
  IF NOT has_old THEN RAISE EXCEPTION 'url_shape key LOST by the seed — carry-forward broken'; END IF;
  IF n_pages <> 27 THEN RAISE EXCEPTION 'pages list changed: expected 27 entries, found %', n_pages; END IF;
END $$;

COMMIT;
