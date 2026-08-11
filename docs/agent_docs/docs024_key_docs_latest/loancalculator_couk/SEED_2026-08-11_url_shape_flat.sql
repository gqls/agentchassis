-- SEED — loancalculator.co.uk: url_shape:"flat" into the structure spec aspect
-- (bugs_open/241 plumbing, BLD-018; framework-rebuild handoff step 3).
--
-- Read by siteUsesFlatURLs (site_url_shape.go), live in the chassis since
-- v1.0.1288 (artefact-verified 2026-08-11: literal present on both replicas,
-- near-miss control absent). With this key, both canonicalisation surfaces emit
-- the dir-flat URL shape (/tools/<s>.html, /guides/<s>.html) this site already
-- serves — measured 2026-08-11: all 24 active tool/guide pages are dir-flat.
--
-- Supersede-then-insert, per convention. The new row carries the OLD data plus
-- the one new key — the structure aspect also holds pages/source/adopted_from,
-- which must survive.
--
-- ⚠ Do NOT rely on `pinned` to protect this row: write_site_spec ignores and
--   drops pinned (LANDMINES). And until the adoption-preservation fix ships,
--   a RE-ADOPTION of this site silently drops this key (LANDMINES 2026-08-11,
--   "Re-adopting a site silently drops the structure spec's opt-in flags").
--   Check after any adoption run:
--     SELECT data ? 'url_shape' FROM site_specs
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
  data || '{"url_shape": "flat"}'::jsonb,
  'operator',
  'url_shape:flat seeded 2026-08-11 (bugs_open/241, BLD-018): site serves dir-flat tool/guide URLs; planner must not move them. Carries forward the adoption-written pages/source/adopted_from unchanged.',
  true,
  'loancalculator_rebuild_thread'
FROM site_specs
WHERE site_id = '0162cde4-633e-45e9-8ca6-87a6b2fe1d26'
  AND aspect = 'structure'
ORDER BY created_at DESC
LIMIT 1;

-- Verify or abort: exactly one current row, key present, pages list intact.
DO $$
DECLARE n int; has_key boolean; n_pages int;
BEGIN
  SELECT count(*) INTO n FROM site_specs
  WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='structure' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION 'expected exactly 1 current structure spec, found %', n; END IF;

  SELECT data ? 'url_shape', jsonb_array_length(data->'pages') INTO has_key, n_pages
  FROM site_specs
  WHERE site_id='0162cde4-633e-45e9-8ca6-87a6b2fe1d26' AND aspect='structure' AND is_current;
  IF NOT has_key THEN RAISE EXCEPTION 'url_shape key absent after seed'; END IF;
  -- 27, not 26: the adoption-written list includes the archived standard-calc.
  -- First run of this seed aborted here expecting 26 (the ACTIVE page count) —
  -- the guard was right to fire, the expectation was wrong. 27 measured:
  -- SELECT jsonb_array_length(data->'pages') on the pre-seed current row.
  IF n_pages <> 27 THEN RAISE EXCEPTION 'pages list changed: expected 27 entries, found %', n_pages; END IF;
END $$;

COMMIT;
