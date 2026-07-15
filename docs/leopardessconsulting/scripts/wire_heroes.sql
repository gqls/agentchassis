-- Wire per-page heroes for leopardess (turn 17). Run ONLY when the assets are active:
--   SELECT asset_key, status FROM assets WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND asset_key IN ('hero_who_we_help','hero_how_we_work');
-- Minimal plan header — NOT a build-site-planner run; content untouched.
DO $$
DECLARE pid uuid;
BEGIN
  SELECT id INTO pid FROM site_plans WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND is_current;
  IF pid IS NULL THEN
    INSERT INTO site_plans (site_id, is_current, source_agent, created_by, notes)
    VALUES ('4851f6fc-71cf-4160-a270-e03d6d3e0732', true, 'operator-rebuild', 'operator-rebuild',
            'Minimal plan header created manually (RUNNING_NOTES turn 17) to carry per-page hero imagery rows. No build-site-planner run — the fixed copy is protected.')
    RETURNING id INTO pid;
  END IF;

  INSERT INTO site_plan_imagery (plan_id, scope, scope_ref, key, kind, prompt, ordering, source)
  SELECT pid, 'page', 'who-we-help', 'hero_who_we_help', 'hero',
         'Flat vector, near-black charcoal ground, hairline antique-gold clusters of connected nodes left/right, calm empty centre. (Generated turn 17, Banana illustration route.)', 0, 'manual'
  WHERE EXISTS (SELECT 1 FROM assets WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND asset_key='hero_who_we_help' AND status='active')
    AND NOT EXISTS (SELECT 1 FROM site_plan_imagery WHERE plan_id=pid AND key='hero_who_we_help');

  INSERT INTO site_plan_imagery (plan_id, scope, scope_ref, key, kind, prompt, ordering, source)
  SELECT pid, 'page', 'how-we-work', 'hero_how_we_work', 'hero',
         'Flat vector, near-black charcoal ground, single hairline gold pipeline with four waypoints and a human checkpoint, calm empty centre. (Generated turn 17.)', 0, 'manual'
  WHERE EXISTS (SELECT 1 FROM assets WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND asset_key='hero_how_we_work' AND status='active')
    AND NOT EXISTS (SELECT 1 FROM site_plan_imagery WHERE plan_id=pid AND key='hero_how_we_work');
END $$;

SELECT sp.is_current, spi.scope, spi.scope_ref, spi.key, spi.kind
FROM site_plans sp LEFT JOIN site_plan_imagery spi ON spi.plan_id=sp.id
WHERE sp.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732';
