-- W7b verify (read-only): the rows, the items draining, the assets landing.
SELECT spi.scope_ref, spi.key, spi.kind, spi.source
FROM site_plan_imagery spi JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
WHERE sp.site_id = (SELECT id FROM sites WHERE domain='idea.uk') AND spi.kind = 'illustration';

SELECT item_key, status, updated_at
FROM site_work_items
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND item_key LIKE 'needs_imagery:section:%illustration%'
ORDER BY item_key;

SELECT asset_key, status, left(url, 60) AS url_head, created_at
FROM assets
WHERE site_id = (SELECT id FROM sites WHERE domain='idea.uk')
  AND asset_key IN ('illustration_home','illustration_tools');
