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