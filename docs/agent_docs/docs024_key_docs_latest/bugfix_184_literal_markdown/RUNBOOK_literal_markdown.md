# RUNBOOK — bugfix 184 (literal markdown)

DB access:
```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## State of the item queue (the bug's live footprint)
```sql
SELECT status, count(*), max(created_at)::date FROM site_work_items
 WHERE item_type='literal_markdown' GROUP BY status ORDER BY 2 DESC;
```
Gotcha: `detected` rows are NOT waiting on triage — they are parked by migration 444's
promoter success floor (`literal_markdown → page-build-handler` is 1/28 lifetime, held until
the ratio recovers past 25%; ~9 hand-promoted successes unholds it). See bugs_open/184's
2026-08-17 CONSUMER NOTICE.

## What forms of markdown are actually live (findings breakdown)
```sql
SELECT f->>'pattern', f->>'source', count(*)
  FROM site_work_items swi, jsonb_array_elements(swi.spec->'findings') f
 WHERE swi.item_type='literal_markdown' AND swi.status NOT IN ('complete','cancelled','rejected')
 GROUP BY 1,2 ORDER BY count DESC;
```
Gotcha: this reads the *check's* findings, so it can only ever show the three patterns the
check knows (bold/code_span/heading). To see forms the check misses, scan content_data raw:
```sql
SELECT count(*) FILTER (WHERE pc.content_data::text ~ '\[[^\]]{2,60}\]\(https?://[^)]{4,120}\)') AS md_link
  FROM page_components pc;
```
(content_data::text is the JSON text — a newline inside a value appears as the two
characters `\n`, so line-anchored regexes must match `\\n` not `^`.)

## Prompt rule / check enablement (are 303/304 still live?)
```sql
SELECT (default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}')
       LIKE '%Plain string also means NO markdown syntax%' AS rule9_extended
  FROM agent_definitions WHERE type='page-content-writer' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
SELECT default_config #> '{workflow,steps,run_checks,config,checks}' ? 'literal_markdown'
  FROM agent_definitions WHERE type='quality-discovery-agent' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Is another session working this bug?
```bash
python3 scripts/who-owns.py 184   # AMBIGUOUS number — the open case is the llm_markdown slug
grep -c literal_markdown ~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl | grep -v ':0'
```
Gotcha: who-owns reads COMMITS only; the transcript grep is what sees an uncommitted session.
A hit can be loaded context (MEMORY/LANDMINES), so read the hit lines before concluding.
