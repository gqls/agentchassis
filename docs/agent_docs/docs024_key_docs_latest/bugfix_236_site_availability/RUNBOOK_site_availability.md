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

`docs/agent_docs/sql_for_agents/372_site_availability_driver_HOLD.sql` — rename
to drop `_HOLD`, then apply via the migration runner, ONLY after:

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | while read p; do
  kubectl -n ai-persona-system exec "$p" -- sh -c 'strings /app/agent-chassis | grep -c "site_unreachable"'
done   # expect >0 on EVERY replica
```

## The break-it-on-purpose drill (both halves) — run 2026-08-11

### A. Safe half: induce a TRUE finding on a pool site

All 17 pool domains are `*.internal` and do not resolve (`curl` exit 6), so this
files a genuine `transport_error` on a domain no visitor can reach.

```sql
BEGIN;
-- Pre-stamp EVERY OTHER agent_type on the rotation table. This is a hard WHERE
-- exclusion for 7 days (COALESCE(last_selected_at,'-infinity') < now() - 7 days),
-- not a sort-order nudge.
INSERT INTO site_discovery_rotation (site_id, agent_type, last_selected_at)
SELECT s.id, a, now() FROM sites s,
  unnest(ARRAY['quality-discovery-agent','design-discovery-agent',
               'completeness-discovery-agent','render-audit-agent']) a
 WHERE s.domain='pool-web-tech.internal'
ON CONFLICT (site_id, agent_type) DO UPDATE SET last_selected_at=EXCLUDED.last_selected_at;
INSERT INTO site_discovery_rotation (site_id, agent_type, last_selected_at)
SELECT s.id,'availability-discovery-agent', now() - interval '30 days' FROM sites s
 WHERE s.domain='pool-web-tech.internal'
ON CONFLICT (site_id, agent_type) DO UPDATE SET last_selected_at=EXCLUDED.last_selected_at;
UPDATE sites SET status='active' WHERE domain='pool-web-tech.internal';  -- flip LAST
COMMIT;
```

**Gotcha 1 — get the agent_type list from the DATA, not from `346_*.sql`.**
`SELECT DISTINCT agent_type FROM site_discovery_rotation` returns **five**; the
migration that created the table names three. `render-audit-agent` was added later
by `369_*.sql` and is the only one of the four currently **enabled**, so an
unstamped site sorts FIRST (`NULLS FIRST`) and gets picked within the hour.

**Gotcha 2 — the pool variant CANNOT prove the self-clear.** `Run()` returns early
when `sites.status` is not `active|deployed`, so reverting the site is exactly what
stops the check probing it. Cancel the item with provenance instead:

```sql
UPDATE site_work_items wi SET status='cancelled',
  spec = spec || jsonb_build_object('drill','…','drill_note','INDUCED, not a real outage: …')
  FROM sites s WHERE s.id=wi.site_id AND s.domain='pool-web-tech.internal'
   AND wi.item_type='site_unreachable' AND wi.status='detected';
DELETE FROM site_discovery_rotation r USING sites s
 WHERE s.id=r.site_id AND s.domain='pool-web-tech.internal';   -- leave as found
UPDATE sites SET status='pool' WHERE domain='pool-web-tech.internal';
```

### B. Real half: take a live site down, both halves, bounded

Script: `bugfix_236_site_availability/` sibling in scratch; the shape that matters —
**time the deletion to land ~25s before the tick**, restore on detect, hard ceiling,
and an `EXIT` trap so a crash still restores.

```bash
# time to the next tick, so the site is not down merely waiting for one
SELECT GREATEST(0, extract(epoch from (last_triggered_at + interval '300 seconds' - now()))::int - 25)
  FROM scheduled_tasks WHERE name='site-discovery-rotation-availability';
```

**Prove BOTH API verbs on a throwaway pattern before touching the real route** —
the token's scope is not guessable from its name (it reads worker routes but NOT
DNS):

```bash
. ~/.cloudflare/404-token.env
API="https://api.cloudflare.com/client/v4/zones/<zone>/workers/routes"
curl -s -X POST "$API" -H "Authorization: Bearer $CLOUDFLARE_API_404_TOKEN" \
  -H 'Content-Type: application/json' \
  --data '{"pattern":"<domain>/__drill-probe-9f3a/*","script":"portfolio-sites-router"}'
curl -s -X DELETE "$API/<id>" -H "Authorization: Bearer $CLOUDFLARE_API_404_TOKEN"
```

**Gotcha 3 — a single `curl` cannot confirm either edge.** Measured: the apex
answered **200 for ~30s after a successful DELETE** and **522 for ~18s after a
successful CREATE**. Poll six times at 5s, and re-list the routes; the API's
`"success": true` is proof the call worked, never that the site's state changed.

**Gotcha 4 — a missing route HANGS the apex, it does not 522 fast.** The probe
recorded `context deadline exceeded` at its 15s timeout, not `http_522`.

### Timing a check — do NOT use its own work item

`site_work_items.created_at` defaults to `now()` = `transaction_timestamp()`, and
`run_discovery_checks` opens ONE transaction (`discovery_checks.go:137`) before the
first check and commits at `:286`. So the item is stamped when the RUN opened.

```sql
-- right: the orchestration brackets the real work, and needs a healthy-run control
SELECT orchestration_id, status, created_at, updated_at, (updated_at-created_at) AS wall
  FROM orchestration_states
 WHERE collected_data->'discovery_result'->>'domain' = '<domain>'
 ORDER BY created_at DESC LIMIT 5;   -- unreachable ~7.6s vs healthy ~1.8s: the delta IS the 5s retry
```

Also: `created_by` is the SENDER's agent type — `generic` for anything the
scheduler dispatches, never `availability-discovery-agent`. `spec->>'check'` names
the producer.

### Reading the pod log for a check — expect a ~20-second buffer

`kubectl logs --since=20m` on a chassis pod returned **20 seconds** of lines while
council seats were running (their DEBUG payload dumps rotate it). Find the right
pod from `orchestration_states.processing_node`, and **run a positive control**
(grep the run's own `orchestration_id`) before treating any absence as evidence.
