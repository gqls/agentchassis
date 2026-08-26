# RUNBOOK — bugfix_410_feed_phase_lock

DB access (all queries below run through this):
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Read the live due machinery (schema first — the column is `interval_seconds`, not `schedule*`)
```sql
\d scheduled_tasks
SELECT name, interval_seconds, enabled, last_triggered_at FROM scheduled_tasks WHERE name='content-feed-refresh';
SELECT default_config->'workflow'->'steps'->'find_news_sites'->'config'->>'query'
FROM agent_definitions WHERE type='content-feed-trigger' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Trigger fire times + per-site orchestrator runs (⚠ orchestration_states prunes at ~2 days)
```sql
SELECT created_at FROM orchestration_states WHERE owner_agent_type='content-feed-trigger' ORDER BY created_at DESC LIMIT 6;
SELECT s.domain, to_char(o.created_at,'DD HH24:MI:SS') FROM orchestration_states o JOIN sites s ON s.id=o.site_id
WHERE o.owner_agent_type='content-feed-orchestrator' AND o.created_at > now()-interval '24 hours' ORDER BY o.created_at;
```

## Per-source cadence truth (the artefact, not the status)
```sql
SELECT s.domain, cs.fetch_interval, cs.next_fetch_at, cs.last_fetched_at, cs.error_count
FROM content_sources cs JOIN sites s ON s.id=cs.site_id WHERE cs.is_active ORDER BY s.domain, cs.next_fetch_at;
```

## Deploy sequence (ORDER IS THE POINT — Go first, config second)
1. Commit is in; ride the next chassis roll (or ask the owner — releases are whole-fleet).
2. Prove the roll carries the commit (per SERVICE; startup line scrolls — fall back to binary probe with a known-sha control):
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <this lane's commit sha> <the stamp>   # exit 0 = shipped
```
3. Only then apply, BY HAND (it is `_HOLD` — the runner deliberately skips it):
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/653_content_feed_due_lookahead_HOLD.sql
```
   Guards abort it if (a) the cadence is no longer 21600 s (fallback literal stale) or
   (b) the live query is not 556's post-image (someone else changed it — merge, don't force).

## Post-fix acceptance (48 h after both halves live)
```sql
-- every 6h-only site: FOUR distinct run-hours/day (≈02:4x, 08:4x, 14:4x, 20:4x)
SELECT s.domain, string_agg(DISTINCT to_char(o.created_at,'HH24'),',' ORDER BY to_char(o.created_at,'HH24'))
FROM orchestration_states o JOIN sites s ON s.id=o.site_id
WHERE o.owner_agent_type='content-feed-orchestrator' AND o.created_at > now()-interval '24 hours' GROUP BY 1;
-- staleness bound: nothing older than 6h15m while the trigger runs
SELECT s.domain, now()-max(cs.last_fetched_at) AS staleness FROM content_sources cs
JOIN sites s ON s.id=cs.site_id WHERE cs.is_active GROUP BY 1 ORDER BY 2 DESC;
```
Gotchas: exclude remortgagecalculator.uk's 2026-08-26 13:43Z off-cadence run from any
before/after census; cap hits (LCO-009, --capped-schedule-ordering) become NORMAL post-fix
(~12 due vs cap 10) — expected demand, not a regression.

## 090 artifacts for this lane's run
```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '15d56c13-2081-431a-ad70-9516c5fcfbc7';
SELECT body FROM doc_notes WHERE body LIKE '%15d56c13%' ORDER BY created_at DESC LIMIT 3;
```
