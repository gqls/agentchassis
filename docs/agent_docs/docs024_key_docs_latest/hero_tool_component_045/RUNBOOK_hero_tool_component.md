# RUNBOOK — bugs_open/045 hero-tool component

`PG` = `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## Selector truth (mirror of component_selector.go queryCandidates)
A `hero-tool` section resolves by `section_type`, NOT by `function`/`name`:
```sql
SELECT function, name FROM content_components
WHERE section_type='hero-tool' AND component_level='section'
  AND is_active=true AND forked_from IS NULL
ORDER BY score DESC;   -- caller takes row 0; NO score threshold
```
So: to change what `hero-tool` resolves to, change **which rows carry
`section_type='hero-tool'`** — not `function`. (023 R2: `slot_name` ↔ `function`;
the *selector* keys on `section_type`.)

## Score formula (for reasoning about ties, not needed once sole candidate)
```
(site_type match ? 0.35 : 0.05) + (page_type match ? 0.15 : 0.0)
+ COALESCE(avg_quality_score,0.3)*0.3
+ (jsonb_array_length(suitable_site_types) BETWEEN 1 AND 3 ? 0.1 : 0.02)
+ LEAST(usage_count/50.0, 1.0)*0.1
```
Bayesian row scored ~0.20 on any real page (empty site_types, page_types
`["bayesian-ranking"]`). Generic scores 0.69 on a tool page.

## Apply the migration (DB config — live immediately, records itself)
```bash
PG < docs/agent_docs/sql_for_agents/183_generic_hero_tool_component.sql
```
It is a single transaction with needle-gates (aborts if the library drifted) and
post-conditions (aborts if the result is wrong). Snapshot table
`bak_bayesian_hero_tool_20260721` is left for rollback.

## Verify the library state
```sql
SELECT name, function, section_type, is_active FROM content_components
WHERE section_type IN ('hero-tool','bayesian-ranking-hero-tool') ORDER BY section_type;
-- expect:  bayesian-ranking-hero-tool_pre_037 | ... | bayesian-ranking-hero-tool | t
--          hero-tool                          | hero-tool | hero-tool           | t
```

## Find every page that requests hero-tool
```sql
SELECT s.domain, p.name, p.page_type, p.build_status, p.sections
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.sections::text LIKE '%hero-tool%' ORDER BY s.domain, p.name;
```

## Current placements of the Bayesian component (deployed heroes)
```sql
SELECT s.domain, p.name, pc.slot_name, pc.position, length(pc.rendered_html)
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.component_id='7cd0408b-5614-4d80-9fe9-9ecddd8ee110' ORDER BY s.domain;
-- 2026-07-21: only gamesdesign.co.uk/bayesian-ranking (correct + deployed).
```

## Trigger a single-page rebuild (verification) — gamesdesign runbook pattern
Re-open the existing page_rerender item (its spec is the shape dispatch expects),
OR flag the page and let discovery create the item:
```sql
UPDATE pages SET build_status='needs_rebuild', built_from_plan_version=NULL, updated_at=now()
WHERE site_id=(SELECT id FROM sites WHERE domain='finetuning.uk') AND name='ai-agent-roi-estimator';
```
The build-dispatch loop claims `approved`/`triaged` items on a cycle; dispatch
latency can be tens of minutes under backlog. Watch:
```sql
SELECT owner_agent_type, status, current_step, created_at FROM orchestration_states
WHERE owner_agent_type IN ('page-build-handler','page-content-writer')
  AND created_at > now() - interval '30 minutes' ORDER BY created_at DESC;
```

## Verify a rebuilt hero is generic (the definitive test)
```sql
-- placement should now be the generic component, no Bayesian strings
SELECT pc.slot_name, pc.component_id::text,
       (pc.rendered_html ~* '(Bayesian|Ranking Free|Calculate Rankings)') AS has_bayes
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='finetuning.uk' AND p.name='ai-agent-roi-estimator'
  AND pc.slot_name IN ('hero-tool');
-- expect: component_id=0bf81196-... , has_bayes=false
```
Then the LIVE page (trust the rendered artefact, not the DB row):
```bash
curl -s https://finetuning.uk/tools/ai-agent-roi-estimator/ | \
  grep -ciE 'Start Ranking Free|Calculate Rankings|Bayesian'   # expect 0
```

## Rollback
```sql
DELETE FROM content_components WHERE name='hero-tool';
UPDATE content_components SET section_type='hero-tool'
  WHERE id='7cd0408b-5614-4d80-9fe9-9ecddd8ee110';
```
