# RUNBOOK — news feed pooling

Commands that were hard to get right, with the gotcha attached.

## DB access

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Reading a live agent definition's workflow

**GOTCHA — this cost the most time.** `agent_definitions` has three workflow
columns (`task_workflow`, `orchestrator_workflow`, `orchestration_workflow`) and
for `content-feed-trigger` **all three are NULL**. The live workflow is in
`default_config->'workflow'`. Querying the workflow columns returns nothing and
looks like the agent is unconfigured.

```sql
SELECT jsonb_pretty(default_config) FROM agent_definitions WHERE type='content-feed-trigger';
```

Also: `agent_definitions` keys on `type`, not `name`, and some types appear on
more than one row — filter `is_active` and check `status` (this one is
`experimental`).

Extract just the site-selection query:

```sql
SELECT jsonb_pretty(default_config) FROM agent_definitions WHERE type='content-feed-trigger';
-- then grep the output for LIMIT / recommended
```

## Counting sites that actually want a news feed

**GOTCHA — `site_specs` is versioned.** A naive count over
`aspect='classification'` returns **1,187 rows across only 11 distinct sites**.
Reading that as "1,176 sites want news" is wrong by two and a half orders of
magnitude. The live trigger filters `ss.is_current = true`; any count that
repeats a figure from that table must do the same.

```sql
SELECT count(*) FROM site_specs ss
 WHERE ss.aspect='classification' AND ss.is_current=true
   AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true;
-- 4, as at 2026-07-19
```

The full live selection query (from `default_config`, verbatim) — note the
`LIMIT 5`, which is the fleet-wide throttle:

```sql
SELECT DISTINCT s.id::text AS site_id, s.domain
  FROM sites s
  JOIN site_specs ss ON ss.site_id = s.id
   AND ss.aspect='classification' AND ss.is_current = true
   AND (ss.data->'content_features'->'news_feed'->>'recommended')::boolean = true
 WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id=s.id AND p.build_status='deployed')
   AND (NOT EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id=s.id AND cs.is_active=true)
     OR EXISTS (SELECT 1 FROM content_sources cs WHERE cs.site_id=s.id AND cs.is_active=true
                 AND (cs.next_fetch_at IS NULL OR cs.next_fetch_at <= NOW())))
 ORDER BY s.domain LIMIT 5;
```

## Current feed footprint

```sql
SELECT source_type, count(*) FROM content_sources GROUP BY 1 ORDER BY 2 DESC;
SELECT s.domain, count(cs.id) AS sources
  FROM sites s LEFT JOIN content_sources cs ON cs.site_id=s.id
 GROUP BY 1 ORDER BY 2 DESC;
SELECT count(*) AS items, count(DISTINCT site_id) AS sites FROM content_feed_items;
```

## Confirming the pooling-relevant schema facts

```sql
\d content_feed_items
```

Two facts the design leans on, both confirmed live 2026-07-19:
- `site_id` is **nullable** (no NOT NULL) — the slot for unassigned/pooled items
  already exists and nothing currently writes NULL.
- `idx_cfi_dedup` is `btree (source_url) WHERE status <> ALL (...)` — **non-unique**,
  so identical articles coexist across sites by design today.

## Checking extensions before designing around vectors

```sql
SELECT extname, extversion FROM pg_extension ORDER BY 1;
-- vector 0.8.0 present as at 2026-07-19
```

## Domain-list analysis

Scratch analysis lives outside the repo (session scratchpad); the method is in
NOTES. Reproduce with a tab-separated `domain<TAB>views` file and the ordered
regex classifier — first rule wins, `misc-brandable` catches the remainder.
**GOTCHA:** naive substring matching gives false positives that matter at these
volumes — `carbondioxide` contains `bond`, `childrensportraits` contains `sport`,
`book-air-taxi` contains `tax`. Anchor or word-boundary the short patterns before
trusting any single bucket's count.
