# RUNBOOK — bugs_open/285 shared-template write

All SQL via `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.

## Is the fix live? (per SERVICE)
Pod start vs commit time is decisive when the image predates the commit:
```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{range .items[*]}{.metadata.name} {.status.startTime} {.spec.containers[0].image}{"\n"}{end}'
git log -1 --format='%ci %h' d7b2d9994
```
Binary probe (`grep -aq <sha> /proc/1/exe`) — use a real-but-different sha as the absent
control; a 40-zero string is PRESENT in the binary and cannot fail.
`kubectl logs … | grep -m1 'build provenance'` on the chassis matched a council-gate payload
line and dumped 3 MB — resolve the pod and read the startup lines, don't `-l` grep a busy pod.

## Casualty (seed 431) — verify at the served page
```sql
SELECT status, attempt_count, updated_at FROM site_work_items
 WHERE item_key='page_rerender:learn-ai-builders-content-first:285-archive-restore';
SELECT length(rendered_html), rendered_html_digest,
       encode(sha256(convert_to(rendered_html,'UTF8')),'hex')=content_data->>'sha256' AS sha_ok
  FROM page_components WHERE id='ff0404b0-f52a-41db-b04a-bc563c2a3a4f';   -- 3781 | NULL | t
```
```bash
curl -sL https://webdesign.co.uk/learn/ai-builders/content-first.html > /tmp/cf.html
grep -c portedPageAssetList /tmp/cf.html            # want 0 (was 1)
grep -c content-first-checklist.pdf /tmp/cf.html    # want 0
grep -c 'class="article-content"' /tmp/cf.html      # want ≥1
```
Fleet fingerprint sweep (positive control = the row above BEFORE the restore):
```sql
SELECT s.domain,p.name FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE pc.rendered_html LIKE '%portedPageAssetList%';
```

## Third-firing watch (until the roll)
```sql
SELECT version_number, created_at, changed_by, length(html_template)
  FROM component_versions WHERE component_id='a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef' ORDER BY 1; -- expect 4 rows
SELECT item_type,status,count(*) FROM site_work_items
 WHERE spec->>'component_id'='a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef' AND created_at > '2026-08-15 17:00'
 GROUP BY 1,2;                                                             -- no new improve_tool
```

## PLAN census (doc_plans, NOT doc_notes)
```sql
WITH ported AS (SELECT regexp_replace(p.name,'^tool-','') k FROM pages p JOIN page_components pc ON pc.page_id=p.id
  WHERE p.site_id=(SELECT id FROM sites WHERE domain='webdesign.co.uk') AND p.page_type='tool'
    AND pc.component_id='a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef')
SELECT count(*) ported_tools, count(dp.id) with_current_plan
  FROM ported LEFT JOIN doc_plans dp ON dp.subject_type='tool' AND dp.subject_key=ported.k AND dp.is_current; -- 63 | 14
```
