# RUNBOOK — bugfix 236 site availability

## Stored index title per deployed site (the 185-safe predicate)

```sql
SELECT s.domain,
       (SELECT p.title FROM pages p
         WHERE p.site_id = s.id AND p.name = 'index'
           AND NOT (p.build_status = 'planned' AND p.deployed_at IS NULL)
         ORDER BY p.updated_at DESC LIMIT 1)
FROM sites s WHERE s.status = 'deployed' ORDER BY s.domain;
```

Gotcha: `pages.status` is `active|archived` and says nothing about deployment
(bugs_open/185). The `NOT (planned AND never-deployed)` shape is
`datahelpers.PageHasShippedPredicateFor` — use the helper in Go, never retype.

## Probe a site the way the check does

```bash
curl -sL --max-time 15 -o /dev/null -w '%{http_code} %{url_effective}\n' \
  -A "agentchassis-discovery/1.0 (+site_unreachable)" "https://<domain>/"
```

Gotcha: `-L` matters (webdesign.uk is a deliberate 302); compare
`%{url_effective}`'s host against the probed domain to spot delegation.

## Is the driver alive and rotating?

```sql
SELECT name, enabled, last_triggered_at FROM scheduled_tasks
 WHERE name = 'site-discovery-rotation-availability';
SELECT r.agent_type, count(*), min(r.last_selected_at), max(r.last_selected_at)
  FROM site_discovery_rotation r
 WHERE r.agent_type = 'availability-discovery-agent' GROUP BY 1;
```

Gotcha: `last_triggered_at` advancing proves only the SCHEDULER ticks
(bugs_open/029's documented trap). Proof of work = a COMPLETED
availability orchestration (24h retention) or an advancing rotation stamp.

## Open/closed availability items

```sql
SELECT wi.status, s.domain, wi.summary, wi.created_at
  FROM site_work_items wi JOIN sites s ON s.id = wi.site_id
 WHERE wi.item_type = 'site_unreachable'
 ORDER BY wi.created_at DESC LIMIT 20;
```

## Force a specific site to be probed next tick

```sql
UPDATE site_discovery_rotation
   SET last_selected_at = now() - interval '30 days'
 WHERE agent_type = 'availability-discovery-agent'
   AND site_id = (SELECT id FROM sites WHERE domain = '<domain>');
```

(No row yet = the site is already first in line — NULLS FIRST.)

## Migration (held until the image rolls)

`docs/agent_docs/sql_for_agents/368_site_availability_driver_HOLD.sql` — rename
to drop `_HOLD`, then apply via the migration runner, ONLY after:

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | while read p; do
  kubectl -n ai-persona-system exec "$p" -- sh -c 'strings /app/agent-chassis | grep -c "site_unreachable"'
done   # expect >0 on EVERY replica
```
