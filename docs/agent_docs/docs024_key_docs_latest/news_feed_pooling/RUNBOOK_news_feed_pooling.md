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

## Migrating a site_specs aspect to a new schema (done for `audience`, 2026-07-20)

**GOTCHA — order is forced by the index.** `idx_site_specs_current` is
`UNIQUE (site_id, aspect) WHERE is_current = true`, so you must supersede the old
row *before* inserting the new one, in one transaction:

```sql
BEGIN;
UPDATE site_specs SET is_current = false, superseded_at = NOW()
 WHERE aspect = '<aspect>' AND is_current = true AND site_id = (SELECT id FROM sites WHERE domain = '<domain>');
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, created_by)
SELECT id, '<aspect>', $j$ {...new shape...} $j$::jsonb,
 'migration', '<workstream>', '<what changed and why, pointing at the PLAN decision>', 'claude/<workstream>'
FROM sites WHERE domain = '<domain>';
COMMIT;
```

- `created_by` is NOT NULL — forgetting it fails the insert.
- Dollar-quote the JSON (`$j$…$j$`): the prose contains apostrophes.
- **Never UPDATE the old row's data in place** — versioning is the rollback.
- Verify by reading back: version chain (`is_current`, `superseded_at`), the new
  row's key set (`jsonb_object_keys`), and **spot-check distinctive phrases** from
  the original so nothing was dropped:
  `SELECT data::text LIKE '%<distinctive phrase>%' FROM site_specs WHERE ...`
- Do NOT invent content the source didn't have — `position` was left null in both
  migrated rows because the source prose contained no intra-portfolio positioning.

The applied one-shot for the two `audience` rows is preserved at
`scratchpad newsfeed/migrate_audience.sql` (session-local) and inline in NOTES.

## The audience.v1 shape (Decision 9)

```json
{
  "schema": "audience.v1",
  "who": {
    "audience_primary":  "<who the reader is>",
    "audience_secondary": "<secondary reader or null>",
    "out_of_scope":      "<who this site is explicitly NOT for, or null>",
    "sophistication":    "technical|professional|casual|luxury|institutional|editorial"
  },
  "position":  "<how this domain differs from OUR sibling domains, or null>",
  "editorial": "<copy/CTA/register directives — content agents only, NEVER ranking>"
}
```

The ranking embedding is computed from `who` + `position` only.

## Pool synthetic sites (created 2026-07-20)

Pools are sites: `pool-<slug>.internal`, **`status='pool'`**, `settings.pool.slug`
for machine identification, `network_id` copied from `system.internal`. One
pool-default `audience.v1` row each.

**Why `status='pool'` and not `'system'`:** fleet loops iterate
`WHERE status='deployed'` (`maintenance_actions.go:694,697`) so any non-deployed
value is excluded — but nothing selects `WHERE status='system'` either
(`system.internal` is referenced only by UUID, `diagnose_triage_action.go:41`),
so a distinct value costs nothing and can never collide with a future
"the system site" convention. `sites` has no CHECK constraints (verified).

List pools:
```sql
SELECT domain, settings->'pool'->>'slug' AS slug FROM sites WHERE status='pool' ORDER BY 1;
```

Safety invariants to re-check after any change touching pools (all must be 0
until pool ingestion is deliberately switched on):
```sql
SELECT count(*) FROM sites WHERE status='deployed' AND domain LIKE 'pool-%';
SELECT count(*) FROM sites s JOIN site_specs ss ON ss.site_id=s.id
  AND ss.aspect='classification' AND ss.is_current WHERE s.status='pool';
```
**GOTCHA:** the second query is the load-bearing one — the content-feed trigger
selects on a current classification spec with `news_feed.recommended=true` plus
a deployed page. Writing a classification spec to a pool site is the act that
arms ingestion; do it knowingly, with sources costed.

The pool creation SQL pattern is in the session scratchpad
(`create_pool_sites.sql`); it is idempotent (`ON CONFLICT DO NOTHING` +
`NOT EXISTS` on the spec insert).

## Domain-list analysis

Scratch analysis lives outside the repo (session scratchpad); the method is in
NOTES. Reproduce with a tab-separated `domain<TAB>views` file and the ordered
regex classifier — first rule wins, `misc-brandable` catches the remainder.
**GOTCHA:** naive substring matching gives false positives that matter at these
volumes — `carbondioxide` contains `bond`, `childrensportraits` contains `sport`,
`book-air-taxi` contains `tax`. Anchor or word-boundary the short patterns before
trusting any single bucket's count.
