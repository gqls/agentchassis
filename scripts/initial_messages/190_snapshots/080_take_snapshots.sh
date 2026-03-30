-- Check which sites to snapshot
SELECT id, domain, status FROM sites WHERE status IN ('deployed', 'published');

-- Snapshot each one
SELECT take_site_snapshot('2a8ebf9c-20a2-4c39-b191-840b012371da', 'manual', NULL, 'Initial baseline', 'admin');
SELECT take_site_snapshot('1368e337-dd1d-4799-bbb3-8221a1b79bcc', 'manual', NULL, 'Initial baseline', 'admin');
SELECT take_site_snapshot('4851f6fc-71cf-4160-a270-e03d6d3e0732', 'manual', NULL, 'Initial baseline', 'admin');

-- Verify
SELECT * FROM v_site_snapshots;

kubectl -n ai-persona-system port-forward svc/core-manager 8088:8088 &

# Then for each site
curl -s -X POST http://localhost:8088/api/v1/admin/sites/2a8ebf9c-20a2-4c39-b191-840b012371da/snapshots \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{"trigger": "manual", "label": "Initial baseline"}'



  -- Pick ai-agent-orchestration.com
  -- Live counts
  SELECT 'pages' AS what, count(*) FROM pages WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  UNION ALL SELECT 'page_components', count(*) FROM page_components WHERE page_id IN (SELECT id FROM pages WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da')
  UNION ALL SELECT 'specs', count(*) FROM site_specs WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da' AND is_current = true
  UNION ALL SELECT 'nav_groups', count(*) FROM site_nav_groups WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  UNION ALL SELECT 'nav_items', count(*) FROM site_nav_items WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  UNION ALL SELECT 'site_components', count(*) FROM site_components WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da';

  -- Snapshot counts for the same site
  SELECT
      jsonb_array_length(spec_snapshot) AS specs,
      jsonb_array_length(pages_snapshot) AS pages,
      jsonb_array_length(nav_snapshot->'groups') AS nav_groups,
      jsonb_array_length(nav_snapshot->'items') AS nav_items,
      jsonb_array_length(components_snapshot) AS site_components,
      (SELECT sum(jsonb_array_length(p->'components')) FROM jsonb_array_elements(pages_snapshot) p) AS page_components
  FROM site_snapshots
  WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da';

---------------------

  If the live and snapshot numbers match across all six categories, the snapshot captured everything.
  and same again for finetuning
  -- finetuning.uk (the biggest one)
  SELECT 'pages', count(*) FROM pages WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc'
  UNION ALL SELECT 'page_components', count(*) FROM page_components WHERE page_id IN (SELECT id FROM pages WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc')
  UNION ALL SELECT 'specs', count(*) FROM site_specs WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc' AND is_current = true;

  SELECT jsonb_array_length(spec_snapshot), jsonb_array_length(pages_snapshot),
      (SELECT sum(jsonb_array_length(p->'components')) FROM jsonb_array_elements(pages_snapshot) p)
  FROM site_snapshots WHERE site_id = '1368e337-dd1d-4799-bbb3-8221a1b79bcc';

