# RUNBOOK — bugs_open/046 truncation casualties

## The census (the bug's own verify query)
```sql
SELECT cc.name, cc.component_level
FROM content_components cc
WHERE cc.is_active = true AND length(cc.html_template) >= 100
  AND (SELECT count(*) FROM regexp_matches(lower(cc.html_template),'<script','g'))
    > (SELECT count(*) FROM regexp_matches(lower(cc.html_template),'</script>','g'));
```
9 at filing (2026-07-20); 8 after the grip-force restore (2026-07-21).

## Full 5-pair predicate (what the discovery check uses; catches exactly the 9)
The check counts open>close for `<script/<style/<section/<div/<fieldset`. Fleet
check that it does NOT over-fire (returned `n_script_only == n_any_pair == 9`):
```sql
WITH c AS (SELECT cc.name,
  (SELECT count(*) FROM regexp_matches(lower(cc.html_template),'<script','g'))
    > (SELECT count(*) FROM regexp_matches(lower(cc.html_template),'</script>','g')) AS ub
  FROM content_components cc WHERE cc.is_active AND length(cc.html_template)>=100)
SELECT count(*) FILTER (WHERE ub) FROM c;
```
GOTCHA: do NOT add `toolTemplateValid`'s ends-mid-token check to the sweep — it
flags 36 legitimate templates fleet-wide. Tag-imbalance alone is precise.

## Per-component triage (on a deployed page? intact prior version?)
```sql
WITH dmg AS (
  SELECT cc.id, cc.name, cc.component_level FROM content_components cc
  WHERE cc.is_active AND length(cc.html_template)>=100
    AND (SELECT count(*) FROM regexp_matches(lower(cc.html_template),'<script','g'))
      > (SELECT count(*) FROM regexp_matches(lower(cc.html_template),'</script>','g')))
SELECT d.name, d.component_level,
  count(DISTINCT p.id) FILTER (WHERE p.status='active' AND p.build_status='deployed') AS deployed_pages,
  count(DISTINCT pc.id) AS page_components
FROM dmg d
LEFT JOIN page_components pc ON pc.component_id=d.id
LEFT JOIN pages p ON p.id=pc.page_id
GROUP BY d.name, d.component_level ORDER BY deployed_pages DESC, d.name;
```
Intact prior version for a component (highest balanced version_number):
```sql
SELECT v.version_number, length(v.html_template),
  (SELECT count(*) FROM regexp_matches(lower(v.html_template),'<script','g'))
    = (SELECT count(*) FROM regexp_matches(lower(v.html_template),'</script>','g')) AS balanced
FROM component_versions v JOIN content_components cc ON cc.id=v.component_id
WHERE cc.name=$1 ORDER BY v.version_number DESC;
```

## Restore a component from an intact version (grip-force pattern — DB, live)
Back up the damaged bytes first, then copy the intact version into the template:
```sql
UPDATE content_components cc SET html_template = v.html_template
FROM component_versions v
WHERE cc.id = '<component_id>' AND v.component_id = '<component_id>'
  AND v.version_number = <intact_version>;
```
This fixes the SOURCE only. The live page updates on the next re-render
(`bugs_open/024`). Verify the row is balanced afterwards (open_s == close_s,
ends `>`). Restoring flips `toolTemplateValid` from reject→accept — desired.

## Build / test the discovery check
```bash
go test ./platform/orchestration/actions/discovery_checks/ -run 'Truncat|Unterminated' -v
```
The package's `TestEveryCheckProducedItemTypeIsClassified` currently also fails on
`contact_form_undeliverable` — a PRE-EXISTING gap from another thread's commit
`3913a0adf`, reproduces without our files, NOT ours to fix.

## Enable the check (image-first — do NOT apply before the image ships)
1. Build + roll a chassis image carrying `check_truncated_component.go`.
2. Confirm it is in the pod:
   `kubectl exec -n ai-persona-system <chassis-pod> -- sh -c 'strings /app/agent-chassis | grep -c truncated_component'`
3. Apply `docs/agent_docs/sql_for_agents/186_enable_truncated_component_check.sql`
   (psql -f + a migration-ledger `--record-only` row the same sitting).

## Live-page state (curl — trust the bytes, not the DB)
```bash
curl -s "https://robot-hands.com/tools/grip-force-friction-calculator/index.html" \
 | grep -c '<script'   # and '</script>' — must match when the page is repaired
```
